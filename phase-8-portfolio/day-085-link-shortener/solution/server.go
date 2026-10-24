package linkshortener

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrNoFreeCode means every generated code was taken, MaxTries times in a row.
var ErrNoFreeCode = errors.New("no free code after several tries")

// MaxBodyBytes is the largest POST /links body the server reads.
const MaxBodyBytes = 4 << 10

// Limiter decides whether a client may create another link now. When it says no,
// it also says how long to wait.
type Limiter interface {
	Allow(key string) (ok bool, retryAfter time.Duration)
}

// IPLimiter gives each key its own token bucket.
type IPLimiter struct {
	mu      sync.Mutex
	every   rate.Limit
	burst   int
	buckets map[string]*rate.Limiter
}

// NewIPLimiter returns a Limiter that lets each key make burst requests at once and
// then r per second.
func NewIPLimiter(r rate.Limit, burst int) *IPLimiter {
	return &IPLimiter{every: r, burst: burst, buckets: map[string]*rate.Limiter{}}
}

// Allow implements Limiter.
func (l *IPLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		b = rate.NewLimiter(l.every, l.burst)
		l.buckets[key] = b
	}
	l.mu.Unlock()

	res := b.Reserve()
	if !res.OK() {
		return false, time.Minute
	}
	if d := res.Delay(); d > 0 {
		res.Cancel() // we won't wait, so give the token back
		return false, d
	}
	return true, 0
}

// Config is everything a Server needs. Store and BaseURL are required; New fills in
// the rest with production defaults.
type Config struct {
	Store    Store
	BaseURL  string               // e.g. "https://kiungo.ke", no trailing slash
	Limiter  Limiter              // default: 10 at once, then 1 per second per client IP
	Lookup   LookupFunc           // default: DefaultLookup
	NewCode  func() string        // default: RandomCode
	Now      func() time.Time     // default: time.Now
	MaxTries int                  // default: 5
	Logf     func(string, ...any) // default: log.Printf
}

// Server is the link shortener's HTTP API.
type Server struct {
	cfg Config
	mux *http.ServeMux
}

// New returns a Server with defaults filled in and its routes registered.
func New(cfg Config) *Server {
	if cfg.Limiter == nil {
		cfg.Limiter = NewIPLimiter(1, 10)
	}
	if cfg.Lookup == nil {
		cfg.Lookup = DefaultLookup
	}
	if cfg.NewCode == nil {
		cfg.NewCode = RandomCode
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.MaxTries <= 0 {
		cfg.MaxTries = 5
	}
	if cfg.Logf == nil {
		cfg.Logf = log.Printf
	}
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	s.mux.HandleFunc("POST /links", s.create)
	s.mux.HandleFunc("GET /links/{code}/stats", s.stats)
	s.mux.HandleFunc("GET /{code}", s.redirect)
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	s.mux.ServeHTTP(w, r)
}

// Shorten validates raw and stores it under a fresh code. When a generated code is
// taken it tries another, up to MaxTries times, then returns ErrNoFreeCode. A bad URL
// gives a *URLError.
func (s *Server) Shorten(ctx context.Context, raw string) (Link, error) {
	u, err := ValidateURL(ctx, raw, s.cfg.Lookup)
	if err != nil {
		return Link{}, err
	}
	for range s.cfg.MaxTries {
		l := Link{Code: s.cfg.NewCode(), URL: u.String(), Created: s.cfg.Now().UTC()}
		err := s.cfg.Store.Create(ctx, l)
		if errors.Is(err, ErrCodeTaken) {
			continue
		}
		if err != nil {
			return Link{}, err
		}
		return l, nil
	}
	return Link{}, ErrNoFreeCode
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Code  string `json:"code"`
	Short string `json:"short"`
	URL   string `json:"url"`
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	if ok, wait := s.cfg.Limiter.Allow(clientIP(r)); !ok {
		secs := int(math.Ceil(wait.Seconds()))
		w.Header().Set("Retry-After", strconv.Itoa(max(secs, 1)))
		writeError(w, http.StatusTooManyRequests, "too many links, slow down")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "body must be JSON like {\"url\":\"https://...\"}")
		return
	}

	l, err := s.Shorten(r.Context(), req.URL)
	var bad *URLError
	switch {
	case errors.As(err, &bad):
		writeError(w, http.StatusUnprocessableEntity, bad.Error())
		return
	case err != nil:
		s.cfg.Logf("POST /links: %v", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	short := s.cfg.BaseURL + "/" + l.Code
	w.Header().Set("Location", short)
	writeJSON(w, http.StatusCreated, createResponse{Code: l.Code, Short: short, URL: l.URL})
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if !ValidCode(code) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	l, err := s.cfg.Store.Hit(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		s.cfg.Logf("GET /%s: %v", code, err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	// 302, not 301: browsers cache a 301 forever and never come back to be counted.
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, l.URL, http.StatusFound)
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if !ValidCode(code) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	l, err := s.cfg.Store.Get(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		s.cfg.Logf("GET /links/%s/stats: %v", code, err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, l)
}

// clientIP is the address the request came from, without the port. Behind a proxy
// you'd read X-Forwarded-For instead, but only from a proxy you trust.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{msg})
}
