// Package observability gives Twende's HTTP API the three things on-call engineers
// need: structured logs with request IDs, Prometheus metrics for every route, and
// health and readiness endpoints.
package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ctxKey is unexported so no other package can collide with it.
type ctxKey struct{}

// RandomID returns 16 random bytes as 32 hex characters.
func RandomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// validID reports whether a client-sent request ID is safe to reuse: 1 to 64 letters,
// digits or hyphens. Anything else could inject fake lines or huge values into logs.
func validID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

// RequestID gives every request an ID. It keeps a valid X-Request-ID header from the
// caller (so one ID follows a request across services) and otherwise makes one with
// newID. The ID goes into the request's context and the response's X-Request-ID header.
func RequestID(newID func() string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !validID(id) {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

// RequestIDFrom returns the request ID stored by RequestID, or "" if there is none.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// recorder remembers the status code a handler wrote. A handler that never calls
// WriteHeader has sent 200.
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

// Logging writes one Info line per request, after it finishes, with the message
// "request" and these attributes: request_id, method, path (r.URL.Path only, never the
// query string, which can hold tokens), status and duration_ms. Requests that answer
// 500 or more are logged at Error level instead. Headers are never logged.
func Logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		level := slog.LevelInfo
		if rec.code() >= 500 {
			level = slog.LevelError
		}
		log.LogAttrs(r.Context(), level, "request",
			slog.String("request_id", RequestIDFrom(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.code()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

// Buckets are the latency histogram's upper bounds in seconds.
var Buckets = []float64{0.05, 0.1, 0.25, 0.5, 1}

// Metrics holds the two RED metrics for an HTTP service: how many requests (and how
// many failed, from the code label) and how long they took.
type Metrics struct {
	requests *prometheus.CounterVec   // http_requests_total{method, route, code}
	duration *prometheus.HistogramVec // http_request_duration_seconds{method, route}
	now      func() time.Time
}

// NewMetrics creates the metrics, registers them with reg and uses now as its clock.
// Pass prometheus.DefaultRegisterer and time.Now in production.
func NewMetrics(reg prometheus.Registerer, now func() time.Time) *Metrics {
	m := &Metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by method, route and status code.",
		}, []string{"method", "route", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "How long HTTP requests took, by method and route.",
			Buckets: Buckets,
		}, []string{"method", "route"}),
		now: now,
	}
	reg.MustRegister(m.requests, m.duration)
	return m
}

// Instrument wraps a handler for one route. route is the pattern, such as
// "/trips/{id}", never the real path: /trips/1, /trips/2 ... would each make a new time
// series and swamp Prometheus. It counts the request with its status code and records
// how long it took.
func (m *Metrics) Instrument(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := m.now()
		rec := &recorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		m.requests.WithLabelValues(r.Method, route, strconv.Itoa(rec.code())).Inc()
		m.duration.WithLabelValues(r.Method, route).Observe(m.now().Sub(start).Seconds())
	})
}

// Live answers the liveness probe: 200 and "ok" while the process can serve HTTP at
// all. It checks nothing else, so a database outage doesn't get every server restarted.
func Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok\n"))
}

// Check is one dependency check for readiness, such as pinging the database.
type Check func(ctx context.Context) error

// Readiness answers the readiness probe: may the load balancer send this server
// traffic? The zero value is not usable; call NewReadiness.
type Readiness struct {
	log      *slog.Logger
	timeout  time.Duration
	mu       sync.Mutex
	checks   map[string]Check
	draining atomic.Bool
}

// NewReadiness returns a Readiness that gives each check timeout to answer.
func NewReadiness(log *slog.Logger, timeout time.Duration) *Readiness {
	return &Readiness{log: log, timeout: timeout, checks: map[string]Check{}}
}

// Add registers a named check.
func (rd *Readiness) Add(name string, c Check) {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.checks[name] = c
}

// Drain marks the server as shutting down: from now on readiness fails, so the load
// balancer stops sending new requests while the ones in flight finish.
func (rd *Readiness) Drain() { rd.draining.Store(true) }

// readyBody is the JSON readiness answers with.
type readyBody struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// ServeHTTP runs every check with a timeout and answers JSON:
// 200 {"status":"ok","checks":{"db":"ok"}} when all pass, or
// 503 {"status":"fail","checks":{"db":"fail"}} when any fails or the server is draining.
// A failing check's error is logged with its name, never sent to the caller, because
// error text can reveal hostnames and versions.
func (rd *Readiness) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rd.mu.Lock()
	names := make([]string, 0, len(rd.checks))
	checks := make(map[string]Check, len(rd.checks)) // a copy, so Add can't race the loop
	for name, c := range rd.checks {
		names = append(names, name)
		checks[name] = c
	}
	rd.mu.Unlock()
	sort.Strings(names)

	body := readyBody{Status: "ok", Checks: map[string]string{}}
	if rd.draining.Load() {
		body.Status = "fail"
	}
	for _, name := range names {
		ctx, cancel := context.WithTimeout(r.Context(), rd.timeout)
		err := checks[name](ctx)
		cancel()
		if err != nil {
			rd.log.Warn("readiness check failed", "check", name, "err", err)
			body.Checks[name] = "fail"
			body.Status = "fail"
		} else {
			body.Checks[name] = "ok"
		}
	}
	code := http.StatusOK
	if body.Status != "ok" {
		code = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}
