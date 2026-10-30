package proxy

import (
	"net/http"
	"net/netip"
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

// Proxy is an http.Handler.
type Proxy struct {
	// Your fields here: the config, the routes (longest prefix first), an
	// *httputil.ReverseProxy, and a mutex-guarded map of *rate.Limiter per client.
}

// New checks cfg and returns a Proxy. Every prefix must start and end with "/" and
// appear once, every upstream must be an absolute http or https URL with a host, and
// Rate and Burst must be above 0.
func New(cfg Config) (*Proxy, error) {
	// Your code here
	return &Proxy{}, nil
}

// ServeHTTP checks the path (400 unless already clean), the allowlist (404), and the
// client's rate limit (429 with Retry-After), then forwards with a timeout. A slow
// upstream is 504, an unreachable one 502, and a body over MaxBody 413.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Your code here
}
