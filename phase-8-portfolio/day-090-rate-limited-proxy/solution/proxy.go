package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Route sends every request whose path starts with Prefix to Upstream, with the
// prefix replaced by the upstream's path: Prefix "/mpesa/" and Upstream
// "http://10.0.0.5:9000/api" send /mpesa/stk/push to http://10.0.0.5:9000/api/stk/push.
type Route struct {
	Prefix   string // starts and ends with "/"
	Upstream string // an absolute http or https URL
}

// Config is the proxy's setup. Routes are the allowlist: nothing else is reachable.
type Config struct {
	Routes         []Route
	Rate           rate.Limit     // requests per second per client
	Burst          int            // requests a client may make at once
	TrustedProxies []netip.Prefix // load balancers whose X-Forwarded-For is believed
	Timeout        time.Duration  // whole request to the upstream; default 10s
	MaxBody        int64          // request body limit in bytes; default 1 MiB
	Transport      http.RoundTripper
	Logf           func(string, ...any)
}

type route struct {
	prefix string
	target *url.URL
}

// Proxy is an http.Handler.
type Proxy struct {
	cfg     Config
	routes  []route // longest prefix first
	rp      *httputil.ReverseProxy
	mu      sync.Mutex
	buckets map[string]*rate.Limiter
}

// New checks cfg and returns a Proxy. Every prefix must start and end with "/" and
// appear once, every upstream must be an absolute http or https URL with a host, and
// Rate and Burst must be above 0.
func New(cfg Config) (*Proxy, error) {
	if len(cfg.Routes) == 0 {
		return nil, errors.New("proxy: at least one route is needed")
	}
	if cfg.Rate <= 0 || cfg.Burst <= 0 {
		return nil, errors.New("proxy: rate and burst must be above 0")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxBody <= 0 {
		cfg.MaxBody = 1 << 20
	}
	if cfg.Logf == nil {
		cfg.Logf = log.Printf
	}
	if cfg.Transport == nil {
		cfg.Transport = &http.Transport{
			Proxy:                 nil, // never pick up HTTP_PROXY from the environment
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: cfg.Timeout,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConnsPerHost:   32,
		}
	}

	p := &Proxy{cfg: cfg, buckets: map[string]*rate.Limiter{}}
	for _, r := range cfg.Routes {
		if !strings.HasPrefix(r.Prefix, "/") || !strings.HasSuffix(r.Prefix, "/") {
			return nil, fmt.Errorf("proxy: prefix %q must start and end with /", r.Prefix)
		}
		u, err := url.Parse(r.Upstream)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, fmt.Errorf("proxy: upstream %q must be an absolute http or https URL", r.Upstream)
		}
		for _, other := range p.routes {
			if other.prefix == r.Prefix {
				return nil, fmt.Errorf("proxy: prefix %q appears twice", r.Prefix)
			}
		}
		p.routes = append(p.routes, route{prefix: r.Prefix, target: u})
	}
	slices.SortFunc(p.routes, func(a, b route) int { return len(b.prefix) - len(a.prefix) })

	p.rp = &httputil.ReverseProxy{
		Rewrite:      p.rewrite,
		Transport:    cfg.Transport,
		ErrorHandler: p.upstreamError,
	}
	return p, nil
}

// cleanPath reports whether p is already in its shortest form: no "..", no ".", no
// doubled slashes. A trailing slash is allowed.
func cleanPath(p string) bool {
	c := path.Clean(p)
	if strings.HasSuffix(p, "/") && c != "/" {
		c += "/"
	}
	return c == p
}

// match finds the route for a cleaned path.
func (p *Proxy) match(urlPath string) (route, bool) {
	for _, r := range p.routes {
		if strings.HasPrefix(urlPath, r.prefix) {
			return r, true
		}
	}
	return route{}, false
}

type routeKey struct{}

// ServeHTTP checks the path (400 unless already clean), the allowlist (404), and the
// client's rate limit (429 with Retry-After), then forwards with a timeout. A slow
// upstream is 504, an unreachable one 502, and a body over MaxBody 413.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /mpesa/../admin must not slip past the prefix check.
	if !cleanPath(r.URL.Path) {
		writeError(w, http.StatusBadRequest, "bad path")
		return
	}
	rt, ok := p.match(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "no such route")
		return
	}
	if ok, wait := p.allow(ClientIP(r, p.cfg.TrustedProxies)); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(wait.Seconds())))))
		writeError(w, http.StatusTooManyRequests, "too many requests")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), p.cfg.Timeout)
	defer cancel()
	ctx = context.WithValue(ctx, routeKey{}, rt)
	r = r.WithContext(ctx)
	r.Body = http.MaxBytesReader(w, r.Body, p.cfg.MaxBody)
	p.rp.ServeHTTP(w, r)
}

// rewrite builds the outbound request. ReverseProxy has already removed hop-by-hop
// and X-Forwarded-* headers from r.Out; StripHopByHop runs again so the rule is visible
// here and can't be lost if this code moves away from ReverseProxy.
func (p *Proxy) rewrite(pr *httputil.ProxyRequest) {
	rt := pr.In.Context().Value(routeKey{}).(route)
	t := rt.target

	rest := strings.TrimPrefix(pr.In.URL.Path, rt.prefix)
	pr.Out.URL.Scheme = t.Scheme
	pr.Out.URL.Host = t.Host
	pr.Out.URL.Path = strings.TrimSuffix(t.Path, "/") + "/" + rest
	pr.Out.URL.RawPath = ""
	pr.Out.URL.RawQuery = pr.In.URL.RawQuery
	pr.Out.Host = t.Host

	StripHopByHop(pr.Out.Header)
	if peer, ok := remoteIP(pr.In); ok && trusted(peer, p.cfg.TrustedProxies) {
		// Keep the chain our own load balancer built; SetXForwarded adds the peer.
		pr.Out.Header["X-Forwarded-For"] = pr.In.Header["X-Forwarded-For"]
	}
	pr.SetXForwarded()
}

// upstreamError answers when the upstream can't be reached or is too slow. The error
// is logged; the client gets a generic message without the upstream's address.
func (p *Proxy) upstreamError(w http.ResponseWriter, r *http.Request, err error) {
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
		writeError(w, http.StatusRequestEntityTooLarge, "body too large")
	case errors.Is(err, context.DeadlineExceeded):
		p.cfg.Logf("proxy: %s %s: upstream timed out: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusGatewayTimeout, "upstream timed out")
	case errors.Is(err, context.Canceled):
		// The client went away. Nobody is left to answer.
	default:
		p.cfg.Logf("proxy: %s %s: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusBadGateway, "upstream unavailable")
	}
}

// allow takes a token from the client's bucket. When the bucket is empty it returns
// false and how long until the next token.
func (p *Proxy) allow(key string) (bool, time.Duration) {
	p.mu.Lock()
	b, ok := p.buckets[key]
	if !ok {
		b = rate.NewLimiter(p.cfg.Rate, p.cfg.Burst)
		p.buckets[key] = b
	}
	p.mu.Unlock()

	res := b.Reserve()
	if d := res.Delay(); d > 0 {
		res.Cancel()
		return false, d
	}
	return true, 0
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{msg})
}
