package corsandsecurityheaders

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestParseOrigins(t *testing.T) {
	got, err := ParseOrigins(" https://twende.example, http://localhost:5173 ,,HTTPS://Admin.Twende.Example ")
	want := []string{"https://twende.example", "http://localhost:5173", "https://admin.twende.example"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ParseOrigins\n got: %q, %v\nwant: %q, <nil>", got, err, want)
	}
	if got, err := ParseOrigins(""); err != nil || len(got) != 0 {
		t.Errorf("ParseOrigins(\"\") = %q, %v; want no origins and no error", got, err)
	}
	for _, bad := range []string{"*", "null", "twende.example", "https://twende.example/", "https://twende.example/app",
		"ftp://twende.example", "https://", "https://twende.example?x=1", "https://user@twende.example"} {
		if _, err := ParseOrigins("https://ok.example," + bad); err == nil {
			t.Errorf("ParseOrigins(%q) error = nil, want an error", bad)
		}
	}
}

func TestLoadOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://twende.example")
	got, err := LoadOrigins()
	if err != nil || !reflect.DeepEqual(got, []string{"https://twende.example"}) {
		t.Errorf("LoadOrigins() = %q, %v", got, err)
	}
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")
	if _, err := LoadOrigins(); err == nil || !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
		t.Errorf("LoadOrigins() with * = %v, want an error naming CORS_ALLOWED_ORIGINS", err)
	}
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "/bookings/9")
	w.WriteHeader(http.StatusCreated)
})

func send(h http.Handler, method, origin string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/bookings", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCORS(t *testing.T) {
	h := CORS([]string{"https://twende.example", "http://localhost:5173"})(okHandler)
	preflight := map[string]string{"Access-Control-Request-Method": "POST", "Access-Control-Request-Headers": "authorization, content-type"}

	tests := []struct {
		name        string
		method      string
		origin      string
		headers     map[string]string
		wantStatus  int
		wantACAO    string
		wantCreds   string
		wantMethods string
	}{
		{"no origin: not a cross-origin call", "POST", "", nil, 201, "", "", ""},
		{"allowed origin", "POST", "https://twende.example", nil, 201, "https://twende.example", "true", ""},
		{"second allowed origin", "POST", "http://localhost:5173", nil, 201, "http://localhost:5173", "true", ""},
		{"other origin: no CORS headers", "POST", "https://evil.example", nil, 201, "", "", ""},
		{"lookalike origin", "POST", "https://twende.example.evil.example", nil, 201, "", "", ""},
		{"preflight from allowed origin", "OPTIONS", "https://twende.example", preflight, 204, "https://twende.example", "true", AllowMethods},
		{"preflight from other origin", "OPTIONS", "https://evil.example", preflight, 403, "", "", ""},
		{"plain OPTIONS is not a preflight", "OPTIONS", "https://twende.example", nil, 201, "https://twende.example", "true", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := send(h, tt.method, tt.origin, tt.headers)
			hd := rec.Header()
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := hd.Get("Access-Control-Allow-Origin"); got != tt.wantACAO {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantACAO)
			}
			if got := hd.Get("Access-Control-Allow-Credentials"); got != tt.wantCreds {
				t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, tt.wantCreds)
			}
			if got := hd.Get("Access-Control-Allow-Methods"); got != tt.wantMethods {
				t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, tt.wantMethods)
			}
			if !slices.Contains(hd.Values("Vary"), "Origin") {
				t.Errorf("Vary = %q, want it to include Origin on every response", hd.Values("Vary"))
			}
			if hd.Get("Access-Control-Allow-Origin") == "*" {
				t.Errorf("Access-Control-Allow-Origin is *: never with credentials")
			}
		})
	}
}

func TestCORSPreflightDetails(t *testing.T) {
	h := CORS([]string{"https://twende.example"})(okHandler)
	rec := send(h, "OPTIONS", "https://twende.example", map[string]string{"Access-Control-Request-Method": "PATCH"})
	hd := rec.Header()
	if got := hd.Get("Access-Control-Allow-Headers"); got != AllowHeaders {
		t.Errorf("Access-Control-Allow-Headers = %q, want %q", got, AllowHeaders)
	}
	if got := hd.Get("Access-Control-Max-Age"); got != MaxAge {
		t.Errorf("Access-Control-Max-Age = %q, want %q", got, MaxAge)
	}
	for _, v := range []string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"} {
		if !slices.Contains(hd.Values("Vary"), v) {
			t.Errorf("preflight Vary = %q, want it to include %s", hd.Values("Vary"), v)
		}
	}
	if hd.Get("Location") != "" {
		t.Errorf("the preflight reached the real handler (Location is set); answer it in the middleware")
	}

	rec = send(h, "POST", "https://twende.example", nil)
	if got := rec.Header().Get("Access-Control-Expose-Headers"); got != ExposeHeaders {
		t.Errorf("Access-Control-Expose-Headers = %q, want %q so the front end can read Location", got, ExposeHeaders)
	}
}

func TestSecurityHeaders(t *testing.T) {
	rec := send(SecurityHeaders(okHandler), "GET", "", nil)
	want := map[string]string{
		"Content-Security-Policy":   "default-src 'none'; frame-ancestors 'none'",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "no-referrer",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if rec.Code != 201 {
		t.Errorf("status = %d, want the handler's 201", rec.Code)
	}
}

func TestCrossOrigin(t *testing.T) {
	h, err := CrossOrigin([]string{"https://twende.example"}, okHandler)
	if err != nil || h == nil {
		t.Fatalf("CrossOrigin: %v", err)
	}
	if _, err := CrossOrigin([]string{"not an origin"}, okHandler); err == nil {
		t.Errorf("CrossOrigin with a bad trusted origin: error = nil, want one")
	}
	tests := []struct {
		name   string
		method string
		origin string
		site   string
		want   int
	}{
		{"curl or a phone app: no browser headers", "POST", "", "", 201},
		{"same-origin browser POST", "POST", "", "same-origin", 201},
		{"cross-site browser POST", "POST", "https://evil.example", "cross-site", 403},
		{"cross-site GET is safe", "GET", "https://evil.example", "cross-site", 201},
		{"trusted front end", "POST", "https://twende.example", "cross-site", 201},
		{"cross-site DELETE", "DELETE", "https://evil.example", "cross-site", 403},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hdr := map[string]string{}
			if tt.site != "" {
				hdr["Sec-Fetch-Site"] = tt.site
			}
			rec := send(h, tt.method, tt.origin, hdr)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
			if tt.want == 403 && rec.Header().Get("Content-Type") != "application/json" {
				t.Errorf("403 Content-Type = %q, want a JSON error", rec.Header().Get("Content-Type"))
			}
		})
	}
}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Add(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.t = c.t.Add(d) }

func TestIPLimiter(t *testing.T) {
	c := &clock{t: time.Date(2026, 3, 2, 17, 0, 0, 0, time.UTC)}
	l := NewIPLimiter(rate.Every(10*time.Second), 3, c.Now) // 3 at once, then 1 every 10 s

	for i := 1; i <= 3; i++ {
		if !l.Allow("41.90.1.1") {
			t.Fatalf("request %d from a new IP refused, want the first 3 allowed", i)
		}
	}
	if l.Allow("41.90.1.1") {
		t.Errorf("request 4 at the same moment allowed, want refused")
	}
	if !l.Allow("41.90.2.2") {
		t.Errorf("another IP was refused: each IP needs its own bucket")
	}
	c.Add(9 * time.Second)
	if l.Allow("41.90.1.1") {
		t.Errorf("allowed after 9 s, want a token only after 10 s")
	}
	c.Add(time.Second)
	if !l.Allow("41.90.1.1") {
		t.Errorf("refused after 10 s, want one new token")
	}
	if got := l.RetryAfter(); got != 10 {
		t.Errorf("RetryAfter() = %d, want 10", got)
	}

	c.Add(5 * time.Minute)
	l.Allow("41.90.3.3")
	if n := l.Evict(5 * time.Minute); n != 2 {
		t.Errorf("Evict(5m) removed %d IPs, want 2 (the two idle ones)", n)
	}
	if n := l.Evict(5 * time.Minute); n != 0 {
		t.Errorf("second Evict removed %d, want 0", n)
	}
}

func TestIPLimiterMiddleware(t *testing.T) {
	c := &clock{t: time.Date(2026, 3, 2, 17, 0, 0, 0, time.UTC)}
	h := NewIPLimiter(rate.Every(30*time.Second), 2, c.Now).Middleware(okHandler)
	req := func(remote, xff string) int {
		r := httptest.NewRequest("POST", "/auth/login", nil)
		r.RemoteAddr = remote
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if rec.Code == 429 && rec.Header().Get("Retry-After") != "30" {
			t.Errorf("429 Retry-After = %q, want \"30\"", rec.Header().Get("Retry-After"))
		}
		return rec.Code
	}
	got := []int{
		req("41.90.1.1:5000", ""),
		req("41.90.1.1:5001", ""),         // a new port is the same client
		req("41.90.1.1:5002", "10.0.0.9"), // a made-up X-Forwarded-For doesn't help
		req("41.90.9.9:5000", ""),
	}
	want := []int{201, 201, 429, 201}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("statuses = %v, want %v", got, want)
	}
}

func TestIPLimiterConcurrent(t *testing.T) {
	l := NewIPLimiter(rate.Every(time.Hour), 5, func() time.Time { return time.Unix(0, 0) })
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("41.90.1.1") {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 5 {
		t.Errorf("50 concurrent requests, %d allowed; want exactly the burst of 5", allowed)
	}
}

func TestWrapOrder(t *testing.T) {
	c := &clock{t: time.Date(2026, 3, 2, 17, 0, 0, 0, time.UTC)}
	h, err := Wrap(okHandler, []string{"https://twende.example"}, NewIPLimiter(rate.Every(time.Minute), 1, c.Now))
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	first := send(h, "POST", "https://twende.example", map[string]string{"Sec-Fetch-Site": "cross-site"})
	if first.Code != 201 {
		t.Fatalf("first request: %d, want 201", first.Code)
	}
	limited := send(h, "POST", "https://twende.example", map[string]string{"Sec-Fetch-Site": "cross-site"})
	if limited.Code != 429 {
		t.Fatalf("second request: %d, want 429", limited.Code)
	}
	if limited.Header().Get("Access-Control-Allow-Origin") != "https://twende.example" {
		t.Errorf("the 429 has no CORS headers, so the browser app can't read it: put CORS outside the limiter")
	}
	if limited.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("the 429 has no security headers: put SecurityHeaders outermost")
	}
}
