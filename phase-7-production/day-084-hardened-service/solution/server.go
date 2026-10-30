// Package hardenedservice is a small Twende booking API made ready for production:
// validated config, timeouts everywhere, rate limiting, security headers, structured
// logs, metrics, health checks and graceful shutdown.
package hardenedservice

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

// Server is the booking API.
type Server struct {
	cfg      Config
	store    Store
	log      *slog.Logger
	reg      *prometheus.Registry
	requests *prometheus.CounterVec
	draining atomic.Bool

	mu       sync.Mutex
	limiters map[string]*rate.Limiter // client IP → its limiter
}

// NewServer returns a Server. Its metrics are registered on reg and served from it.
func NewServer(cfg Config, store Store, log *slog.Logger, reg *prometheus.Registry) *Server {
	s := &Server{
		cfg:      cfg,
		store:    store,
		log:      log,
		reg:      reg,
		limiters: map[string]*rate.Limiter{},
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by method, route and status code.",
		}, []string{"method", "route", "code"}),
	}
	reg.MustRegister(s.requests)
	return s
}

// HTTPServer returns the *http.Server to run Handler with, with every timeout set:
// a server without them can be held open for ever by slow clients.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 16,
		ErrorLog:          slog.NewLogLogger(s.log.Handler(), slog.LevelWarn),
	}
}

// Handler returns every route wrapped in the middleware, outermost first: request ID,
// logging, panic recovery (inside logging, so a recovered panic is still logged as a
// 500), security headers, then rate limiting for the API routes only.
func (s *Server) Handler() http.Handler {
	api := http.NewServeMux()
	api.Handle("GET /trips", s.instrument("/trips", http.HandlerFunc(s.listTrips)))
	api.Handle("POST /bookings", s.instrument("/bookings", http.HandlerFunc(s.createBooking)))

	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", s.healthz)
	root.HandleFunc("GET /readyz", s.readyz)
	root.Handle("GET /metrics", promhttp.HandlerFor(s.reg, promhttp.HandlerOpts{}))
	root.Handle("/", s.rateLimit(api))

	return s.requestID(s.logRequests(s.recoverPanics(securityHeaders(root))))
}

// Run serves on ln until ctx is cancelled, then shuts down gracefully: readiness starts
// failing at once, no new connections are accepted, and requests in flight get up to
// ShutdownGrace to finish. It returns nil after a clean shutdown.
func (s *Server) Run(ctx context.Context, ln net.Listener) error {
	srv := s.HTTPServer()
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	s.draining.Store(true)
	s.log.Info("shutting down", "grace", s.cfg.ShutdownGrace.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// --- middleware ---

type ctxKey struct{}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func (r *recorder) code() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		level := slog.LevelInfo
		if rec.code() >= 500 {
			level = slog.LevelError
		}
		s.log.LogAttrs(r.Context(), level, "request",
			slog.String("request_id", requestIDFrom(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.code()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

// recoverPanics turns a panic into a 500 and an error log line, so one bad request
// can't take the process down. http.ErrAbortHandler is re-panicked, as net/http expects.
func (s *Server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler {
				panic(v)
			}
			s.log.Error("panic", "request_id", requestIDFrom(r.Context()), "panic", v)
			writeError(w, http.StatusInternalServerError, "something went wrong")
		}()
		next.ServeHTTP(w, r)
	})
}

// securityHeaders sets the headers from Day 61 on every response, errors included.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// clientIP is the address the connection came from. X-Forwarded-For is ignored: any
// client can send it. Behind a trusted proxy you would read it from that proxy only.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) limiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.limiters[ip]
	if !ok {
		l = rate.NewLimiter(rate.Limit(s.cfg.RatePerSecond), s.cfg.Burst)
		s.limiters[ip] = l
	}
	return l
}

// rateLimit answers 429 with Retry-After once a client IP has used its burst.
func (s *Server) rateLimit(next http.Handler) http.Handler {
	retry := strconv.Itoa(int(math.Ceil(1 / s.cfg.RatePerSecond)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter(clientIP(r)).Allow() {
			w.Header().Set("Retry-After", retry)
			writeError(w, http.StatusTooManyRequests, "slow down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) instrument(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		s.requests.WithLabelValues(r.Method, route, strconv.Itoa(rec.code())).Inc()
	})
}

// --- handlers ---

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.HandlerTimeout)
	defer cancel()
	trips, err := s.store.ListTrips(ctx)
	if err != nil {
		s.storeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, trips)
}

var phonePattern = regexp.MustCompile(`^0[17][0-9]{8}$`)

type bookingRequest struct {
	TripID string `json:"tripId"`
	Seat   int    `json:"seat"`
	Phone  string `json:"phone"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.APIKey)) != 1 {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "missing or wrong API key")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req bookingRequest
	if err := dec.Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "body must be one JSON object with tripId, seat and phone")
		return
	}
	if req.TripID == "" || !phonePattern.MatchString(req.Phone) {
		writeError(w, http.StatusUnprocessableEntity, "tripId is required and phone must look like 07XXXXXXXX or 01XXXXXXXX")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.HandlerTimeout)
	defer cancel()
	b, err := s.store.Book(ctx, req.TripID, req.Seat, req.Phone)
	switch {
	case errors.Is(err, ErrNoTrip):
		writeError(w, http.StatusNotFound, "no such trip")
	case errors.Is(err, ErrSeatTaken):
		writeError(w, http.StatusConflict, "seat already taken")
	case errors.Is(err, ErrBadSeat):
		writeError(w, http.StatusUnprocessableEntity, "no such seat on this trip")
	case err != nil:
		s.storeError(w, r, err)
	default:
		w.Header().Set("Location", "/bookings/"+b.ID)
		writeJSON(w, http.StatusCreated, b)
	}
}

// storeError logs a store failure and answers 503, the same for a timeout or a broken
// database: the caller may retry, and doesn't learn anything about our internals.
func (s *Server) storeError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("store failed", "request_id", requestIDFrom(r.Context()), "err", err)
	writeError(w, http.StatusServiceUnavailable, "try again shortly")
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok\n"))
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if s.draining.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "draining"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.HandlerTimeout)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.log.Warn("readiness: store ping failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "fail"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
