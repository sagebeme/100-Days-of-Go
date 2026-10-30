package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

var lb = []netip.Prefix{netip.MustParsePrefix("10.0.0.0/24")} // our load balancers

func TestStripHopByHop(t *testing.T) {
	h := http.Header{}
	h.Set("Connection", "keep-alive, X-Debug-Token")
	h.Add("Connection", "X-Other")
	h.Set("Keep-Alive", "timeout=5")
	h.Set("Proxy-Authorization", "Basic c2VjcmV0")
	h.Set("Te", "trailers")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Upgrade", "websocket")
	h.Set("X-Debug-Token", "abc")
	h.Set("X-Other", "1")
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer kept")
	StripHopByHop(h)
	want := http.Header{"Content-Type": {"application/json"}, "Authorization": {"Bearer kept"}}
	if len(h) != len(want) || h.Get("Content-Type") != "application/json" || h.Get("Authorization") != "Bearer kept" {
		t.Errorf("after StripHopByHop, headers = %v, want only %v", h, want)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name, remote, xff, want string
	}{
		{"direct client", "41.90.64.7:5000", "", "41.90.64.7"},
		{"direct client faking XFF", "41.90.64.7:5000", "1.2.3.4", "41.90.64.7"},
		{"via our load balancer", "10.0.0.2:443", "41.90.64.7", "41.90.64.7"},
		{"via LB, client faked a prefix", "10.0.0.2:443", "1.2.3.4, 41.90.64.7", "41.90.64.7"},
		{"via two of our proxies", "10.0.0.2:443", "41.90.64.7, 10.0.0.9", "41.90.64.7"},
		{"via LB with no XFF", "10.0.0.2:443", "", "10.0.0.2"},
		{"via LB with junk", "10.0.0.2:443", "41.90.64.7, not-an-ip", "10.0.0.2"},
		{"IPv6 client", "[2c0f:fe38::1]:5000", "", "2c0f:fe38::1"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tt.remote
		if tt.xff != "" {
			r.Header.Set("X-Forwarded-For", tt.xff)
		}
		if got := ClientIP(r, lb); got != tt.want {
			t.Errorf("%s: ClientIP = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNewRejectsBadConfig(t *testing.T) {
	ok := Route{Prefix: "/mpesa/", Upstream: "http://127.0.0.1:9000"}
	tests := []struct {
		name string
		cfg  Config
	}{
		{"no routes", Config{Rate: 1, Burst: 1}},
		{"no rate", Config{Routes: []Route{ok}, Burst: 1}},
		{"no burst", Config{Routes: []Route{ok}, Rate: 1}},
		{"prefix without slashes", Config{Routes: []Route{{Prefix: "mpesa", Upstream: ok.Upstream}}, Rate: 1, Burst: 1}},
		{"relative upstream", Config{Routes: []Route{{Prefix: "/a/", Upstream: "/local"}}, Rate: 1, Burst: 1}},
		{"file upstream", Config{Routes: []Route{{Prefix: "/a/", Upstream: "file:///etc"}}, Rate: 1, Burst: 1}},
		{"duplicate prefix", Config{Routes: []Route{ok, ok}, Rate: 1, Burst: 1}},
	}
	for _, tt := range tests {
		if _, err := New(tt.cfg); err == nil {
			t.Errorf("%s: New succeeded, want an error", tt.name)
		}
	}
}

// seen is what the upstream echoes back about the request it received.
type seen struct {
	Path, Query, Host, XFF, XFH string
	Headers                     http.Header
}

func echo(t *testing.T) *httptest.Server {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "X-Upstream-Secret")
		w.Header().Set("X-Upstream-Secret", "leak")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seen{
			Path: r.URL.Path, Query: r.URL.RawQuery, Host: r.Host,
			XFF: r.Header.Get("X-Forwarded-For"), XFH: r.Header.Get("X-Forwarded-Host"), Headers: r.Header,
		})
	}))
	t.Cleanup(up.Close)
	return up
}

func newProxy(t *testing.T, cfg Config) *Proxy {
	t.Helper()
	if cfg.Rate == 0 {
		cfg.Rate, cfg.Burst = 1000, 1000
	}
	cfg.TrustedProxies = lb
	cfg.Logf = func(string, ...any) {}
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return p
}

func send(p http.Handler, method, target, remote string, header http.Header) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	r.RemoteAddr = remote
	for k, v := range header {
		r.Header[k] = v
	}
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)
	return w
}

func TestForwarding(t *testing.T) {
	up := echo(t)
	p := newProxy(t, Config{Routes: []Route{
		{Prefix: "/mpesa/", Upstream: up.URL + "/api"},
		{Prefix: "/mpesa/v2/", Upstream: up.URL + "/v2api/"},
		{Prefix: "/", Upstream: up.URL + "/site"},
	}})

	tests := []struct{ target, wantPath, wantQuery string }{
		{"/mpesa/stk/push?amount=100", "/api/stk/push", "amount=100"},
		{"/mpesa/v2/balance", "/v2api/balance", ""},
		{"/about", "/site/about", ""},
	}
	for _, tt := range tests {
		w := send(p, "GET", "http://pesa.example"+tt.target, "41.90.64.7:5000", http.Header{
			"Connection":          {"X-Client-Hop"},
			"X-Client-Hop":        {"drop me"},
			"Proxy-Authorization": {"Basic c2VjcmV0"},
			"X-Forwarded-For":     {"6.6.6.6"},
			"Authorization":       {"Bearer token-kept"},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200; body %s", tt.target, w.Code, w.Body)
		}
		var got seen
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("upstream answer isn't JSON: %v", err)
		}
		if got.Path != tt.wantPath || got.Query != tt.wantQuery {
			t.Errorf("GET %s reached %s?%s, want %s?%s", tt.target, got.Path, got.Query, tt.wantPath, tt.wantQuery)
		}
		if got.XFF != "41.90.64.7" {
			t.Errorf("upstream saw X-Forwarded-For %q, want just the real client 41.90.64.7", got.XFF)
		}
		if got.XFH != "pesa.example" {
			t.Errorf("upstream saw X-Forwarded-Host %q, want pesa.example", got.XFH)
		}
		if got.Host != strings.TrimPrefix(up.URL, "http://") {
			t.Errorf("upstream saw Host %q, want its own host", got.Host)
		}
		for _, h := range []string{"X-Client-Hop", "Proxy-Authorization", "Connection"} {
			if v := got.Headers.Get(h); v != "" {
				t.Errorf("upstream received hop-by-hop header %s: %q", h, v)
			}
		}
		if got.Headers.Get("Authorization") != "Bearer token-kept" {
			t.Errorf("Authorization wasn't forwarded")
		}
		if v := w.Header().Get("X-Upstream-Secret"); v != "" {
			t.Errorf("client received the upstream's hop-by-hop header: %q", v)
		}
	}
}

func TestForwardedChainFromOurLoadBalancer(t *testing.T) {
	up := echo(t)
	p := newProxy(t, Config{Routes: []Route{{Prefix: "/", Upstream: up.URL}}})
	w := send(p, "GET", "/x", "10.0.0.2:443", http.Header{"X-Forwarded-For": {"41.90.64.7"}})
	var got seen
	json.NewDecoder(w.Body).Decode(&got)
	if got.XFF != "41.90.64.7, 10.0.0.2" {
		t.Errorf("X-Forwarded-For = %q, want %q", got.XFF, "41.90.64.7, 10.0.0.2")
	}
}

func TestAllowlistAndPaths(t *testing.T) {
	var hits atomic.Int64
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
	t.Cleanup(up.Close)
	p := newProxy(t, Config{Routes: []Route{{Prefix: "/mpesa/", Upstream: up.URL}}})

	tests := []struct {
		target string
		want   int
	}{
		{"/admin/users", http.StatusNotFound},
		{"/mpesa", http.StatusNotFound},
		{"/mpesa/../admin", http.StatusBadRequest},
		{"/mpesa/./x", http.StatusBadRequest},
		{"/mpesa//x", http.StatusBadRequest},
		{"http://169.254.169.254/latest/meta-data/", http.StatusNotFound},
	}
	for _, tt := range tests {
		if w := send(p, "GET", tt.target, "41.90.64.7:5000", nil); w.Code != tt.want {
			t.Errorf("GET %s status = %d, want %d", tt.target, w.Code, tt.want)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("the upstream was hit %d times, want 0", n)
	}
	if w := send(p, "GET", "/mpesa/ok/", "41.90.64.7:5000", nil); w.Code != http.StatusOK {
		t.Errorf("GET /mpesa/ok/ status = %d, want 200", w.Code)
	}
}

func TestRateLimitPerClient(t *testing.T) {
	up := echo(t)
	p := newProxy(t, Config{Routes: []Route{{Prefix: "/", Upstream: up.URL}}, Rate: rate.Every(time.Hour), Burst: 2})

	for i := 1; i <= 2; i++ {
		if w := send(p, "GET", "/", "41.90.64.7:5000", nil); w.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i, w.Code)
		}
	}
	w := send(p, "GET", "/", "41.90.64.7:6000", nil) // new port, same client
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("3rd request status = %d, want 429", w.Code)
	}
	if ra := w.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want whole seconds above 0", ra)
	}

	// Faking X-Forwarded-For doesn't get a fresh bucket.
	if w := send(p, "GET", "/", "41.90.64.7:7000", http.Header{"X-Forwarded-For": {"8.8.8.8"}}); w.Code != http.StatusTooManyRequests {
		t.Errorf("a faked X-Forwarded-For got status %d, want 429", w.Code)
	}
	// Two clients behind our load balancer have separate buckets.
	for _, client := range []string{"196.201.214.1", "196.201.214.2"} {
		if w := send(p, "GET", "/", "10.0.0.2:443", http.Header{"X-Forwarded-For": {client}}); w.Code != http.StatusOK {
			t.Errorf("client %s behind the LB got %d, want 200", client, w.Code)
		}
	}
}

func TestUpstreamErrors(t *testing.T) {
	release := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(release); slow.Close() })

	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close() // nothing listens there now

	// Two proxies, so the dial error never races a tight timeout: the dead upstream
	// gets plenty of time to fail on its own, and only the slow one can time out.
	slowProxy := newProxy(t, Config{Routes: []Route{{Prefix: "/slow/", Upstream: slow.URL}}, Timeout: 50 * time.Millisecond})
	deadProxy := newProxy(t, Config{Routes: []Route{{Prefix: "/dead/", Upstream: deadURL}}, Timeout: 30 * time.Second})

	w := send(slowProxy, "GET", "/slow/x", "41.90.64.7:5000", nil)
	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("slow upstream status = %d, want 504", w.Code)
	}
	w = send(deadProxy, "GET", "/dead/x", "41.90.64.7:5000", nil)
	if w.Code != http.StatusBadGateway {
		t.Errorf("dead upstream status = %d, want 502", w.Code)
	}
	if strings.Contains(w.Body.String(), "127.0.0.1") {
		t.Errorf("the error leaks the upstream address: %s", w.Body)
	}
}

func TestBodyLimit(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		for {
			if _, err := r.Body.Read(buf); err != nil {
				return
			}
		}
	}))
	t.Cleanup(up.Close)
	p := newProxy(t, Config{Routes: []Route{{Prefix: "/", Upstream: up.URL}}, MaxBody: 10})

	r := httptest.NewRequest("POST", "/upload", strings.NewReader(strings.Repeat("x", 100)))
	r.RemoteAddr = "41.90.64.7:5000"
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("a 100-byte body with MaxBody 10: status = %d, want 413", w.Code)
	}
}
