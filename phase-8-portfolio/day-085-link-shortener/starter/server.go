package linkshortener

import (
	"context"
	"errors"
	"net/http"
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
	// Your fields here: a mutex, the rate, the burst and a map of *rate.Limiter.
}

// NewIPLimiter returns a Limiter that lets each key make burst requests at once and
// then r per second.
func NewIPLimiter(r rate.Limit, burst int) *IPLimiter {
	// Your code here
	return &IPLimiter{}
}

// Allow implements Limiter.
func (l *IPLimiter) Allow(key string) (bool, time.Duration) {
	// Your code here
	return false, 0
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

// New returns a Server with defaults filled in and its routes registered:
// POST /links, GET /{code} and GET /links/{code}/stats.
func New(cfg Config) *Server {
	// Your code here
	return &Server{cfg: cfg, mux: http.NewServeMux()}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Shorten validates raw and stores it under a fresh code. When a generated code is
// taken it tries another, up to MaxTries times, then returns ErrNoFreeCode. A bad URL
// gives a *URLError.
func (s *Server) Shorten(ctx context.Context, raw string) (Link, error) {
	// Your code here
	return Link{}, nil
}
