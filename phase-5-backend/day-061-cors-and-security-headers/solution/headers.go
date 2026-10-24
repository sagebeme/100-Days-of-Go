// Package corsandsecurityheaders wraps the Twende API in the middleware a browser
// front end needs: CORS for the allowed origins only, security headers, cross-origin
// request protection, and a rate limit per IP.
package corsandsecurityheaders

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// ParseOrigins reads a comma-separated list such as
// "https://twende.example, http://localhost:5173". Each origin must be http or https,
// with a host and nothing after it: no path, query, fragment or user. "*" and "null"
// are refused. Origins are returned lower-cased, and blanks between commas are skipped.
func ParseOrigins(raw string) ([]string, error) {
	var out []string
	var errs []error
	for _, part := range strings.Split(raw, ",") {
		o := strings.ToLower(strings.TrimSpace(part))
		if o == "" {
			continue
		}
		u, err := url.Parse(o)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
			u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil || strings.HasSuffix(o, "?") || strings.HasSuffix(o, "#") {
			errs = append(errs, fmt.Errorf("%q is not an origin like https://twende.example", part))
			continue
		}
		out = append(out, o)
	}
	return out, errors.Join(errs...)
}

// LoadOrigins reads CORS_ALLOWED_ORIGINS with ParseOrigins. Unset means no origins.
func LoadOrigins() ([]string, error) {
	origins, err := ParseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS: %w", err)
	}
	return origins, nil
}

// The CORS answers the API gives to an allowed origin.
const (
	AllowMethods  = "GET, POST, PUT, PATCH, DELETE"
	AllowHeaders  = "Authorization, Content-Type, Idempotency-Key"
	ExposeHeaders = "Location, Retry-After"
	MaxAge        = "600"
)

// CORS returns middleware that lets the listed origins call the API from a browser,
// with credentials. It always adds "Vary: Origin". For an allowed origin it echoes
// the origin back (never "*"), allows credentials and exposes ExposeHeaders. A
// preflight (OPTIONS with Access-Control-Request-Method) from an allowed origin gets
// 204 with the Allow headers and Max-Age; from any other origin, 403.
func CORS(allowed []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Add("Vary", "Origin")
			origin := r.Header.Get("Origin")
			preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			if !slices.Contains(allowed, origin) {
				if preflight {
					writeError(w, http.StatusForbidden, "origin not allowed")
					return
				}
				next.ServeHTTP(w, r) // no CORS headers: the browser hides the response
				return
			}
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			if preflight {
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
				h.Set("Access-Control-Allow-Methods", AllowMethods)
				h.Set("Access-Control-Allow-Headers", AllowHeaders)
				h.Set("Access-Control-Max-Age", MaxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.Set("Access-Control-Expose-Headers", ExposeHeaders)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets, on every response, the headers a JSON API should send:
// a CSP that allows nothing and no framing, HSTS for two years including
// subdomains, nosniff, and no Referer.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// CrossOrigin wraps next in http.CrossOriginProtection, which refuses cross-site
// POST, PUT, PATCH and DELETE requests from browsers, except from the trusted
// origins. A refused request gets a JSON 403.
func CrossOrigin(trusted []string, next http.Handler) (http.Handler, error) {
	p := http.NewCrossOriginProtection()
	for _, o := range trusted {
		if err := p.AddTrustedOrigin(o); err != nil {
			return nil, err
		}
	}
	p.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusForbidden, "cross-origin request refused")
	}))
	return p.Handler(next), nil
}

type client struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// IPLimiter keeps one token bucket per client IP. It is safe for concurrent use.
type IPLimiter struct {
	Rate  rate.Limit // tokens added per second
	Burst int        // bucket size
	Now   func() time.Time

	mu      sync.Mutex
	clients map[string]*client
}

// NewIPLimiter returns a limiter that allows each IP burst requests at once and
// r per second after that.
func NewIPLimiter(r rate.Limit, burst int, now func() time.Time) *IPLimiter {
	return &IPLimiter{Rate: r, Burst: burst, Now: now, clients: map[string]*client{}}
}

// Allow reports whether ip may make a request now, and uses up a token if so.
func (l *IPLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.Now()
	c, ok := l.clients[ip]
	if !ok {
		c = &client{lim: rate.NewLimiter(l.Rate, l.Burst)}
		l.clients[ip] = c
	}
	c.lastSeen = now
	return c.lim.AllowN(now, 1)
}

// Evict forgets every IP not seen for idle or longer and returns how many it removed.
// Call it every few minutes so the map doesn't grow forever.
func (l *IPLimiter) Evict(idle time.Duration) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.Now()
	n := 0
	for ip, c := range l.clients {
		if now.Sub(c.lastSeen) >= idle {
			delete(l.clients, ip)
			n++
		}
	}
	return n
}

// RetryAfter is the whole number of seconds until a new token arrives, at least 1.
func (l *IPLimiter) RetryAfter() int {
	if l.Rate <= 0 {
		return 60
	}
	return max(1, int(math.Ceil(1/float64(l.Rate))))
}

// clientIP is the IP from r.RemoteAddr. X-Forwarded-For is ignored: anyone can send it.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Middleware answers 429 with Retry-After when the client's IP is over the limit.
func (l *IPLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.Allow(ip) {
			slog.Warn("rate limited", "ip", ip, "path", r.URL.Path)
			w.Header().Set("Retry-After", strconv.Itoa(l.RetryAfter()))
			writeError(w, http.StatusTooManyRequests, "too many requests, slow down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Wrap puts the API's middleware around mux, outermost first: security headers, so
// every response has them; CORS, so even a 429 or 403 is readable by the front end;
// the rate limit; then cross-origin protection.
func Wrap(mux http.Handler, origins []string, limiter *IPLimiter) (http.Handler, error) {
	protected, err := CrossOrigin(origins, mux)
	if err != nil {
		return nil, err
	}
	return SecurityHeaders(CORS(origins)(limiter.Middleware(protected))), nil
}
