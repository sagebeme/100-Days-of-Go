package hardenedservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const apiKey = "kiosk-key-0123456789abcdef0123456789abcdef"

func getenv(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func validEnv() map[string]string {
	return map[string]string{"TWENDE_API_KEY": apiKey}
}

// syncBuffer is a bytes.Buffer that's safe to write from the server's goroutines.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

var trips = []Trip{
	{ID: "46K-0730", Route: "46K Kencom to Kawangware", Departs: "07:30", Seats: 14},
	{ID: "125-0745", Route: "125 Railways to Ongata Rongai", Departs: "07:45", Seats: 33},
}

// newServer builds a Server from validEnv plus extra, over store (a MemStore if nil).
func newServer(t *testing.T, extra map[string]string, store Store) (*Server, http.Handler, *syncBuffer) {
	t.Helper()
	env := validEnv()
	for k, v := range extra {
		env[k] = v
	}
	cfg, err := LoadConfig(getenv(env))
	if err != nil {
		t.Fatalf("LoadConfig(valid env) = %v", err)
	}
	if store == nil {
		store = NewMemStore(trips...)
	}
	logs := &syncBuffer{}
	s := NewServer(cfg, store, slog.New(slog.NewJSONHandler(logs, nil)), prometheus.NewRegistry())
	return s, s.Handler(), logs
}

func do(h http.Handler, method, path, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func book(h http.Handler, body string) *httptest.ResponseRecorder {
	return do(h, "POST", "/bookings", body, "Authorization", "Bearer "+apiKey, "Content-Type", "application/json")
}

// --- Requirement 1: config is validated at startup ---

func TestConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig(getenv(validEnv()))
	if err != nil {
		t.Fatalf("LoadConfig = %v", err)
	}
	want := Config{
		Addr: "localhost:8080", Env: "dev", APIKey: apiKey, RatePerSecond: 5, Burst: 10,
		HandlerTimeout: 2 * time.Second, ShutdownGrace: 10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second, MaxBodyBytes: 1 << 20,
	}
	if cfg != want {
		t.Errorf("LoadConfig(defaults) =\n%#v\nwant\n%#v", cfg, want)
	}
	env := validEnv()
	env["TWENDE_ENV"], env["TWENDE_RATE"], env["TWENDE_BURST"], env["TWENDE_HANDLER_TIMEOUT"] = "prod", "0.5", "3", "750ms"
	cfg, err = LoadConfig(getenv(env))
	if err != nil || cfg.Env != "prod" || cfg.RatePerSecond != 0.5 || cfg.Burst != 3 || cfg.HandlerTimeout != 750*time.Millisecond {
		t.Errorf("LoadConfig(overrides) = %+v, %v", cfg, err)
	}
}

func TestConfigRejectsBadValues(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		vars []string // every one must be named in the error
	}{
		{"no api key", map[string]string{}, []string{"TWENDE_API_KEY"}},
		{"short api key", map[string]string{"TWENDE_API_KEY": "short"}, []string{"TWENDE_API_KEY"}},
		{"bad env", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_ENV": "production"}, []string{"TWENDE_ENV"}},
		{"rate not a number", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_RATE": "fast"}, []string{"TWENDE_RATE"}},
		{"rate zero", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_RATE": "0"}, []string{"TWENDE_RATE"}},
		{"burst zero", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_BURST": "0"}, []string{"TWENDE_BURST"}},
		{"timeout without unit", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_HANDLER_TIMEOUT": "2"}, []string{"TWENDE_HANDLER_TIMEOUT"}},
		{"negative grace", map[string]string{"TWENDE_API_KEY": apiKey, "TWENDE_SHUTDOWN_GRACE": "-1s"}, []string{"TWENDE_SHUTDOWN_GRACE"}},
		{"everything at once", map[string]string{"TWENDE_ENV": "x", "TWENDE_RATE": "-1", "TWENDE_BURST": "many"}, []string{"TWENDE_API_KEY", "TWENDE_ENV", "TWENDE_RATE", "TWENDE_BURST"}},
	}
	for _, tt := range tests {
		_, err := LoadConfig(getenv(tt.env))
		if err == nil {
			t.Errorf("%s: LoadConfig = nil error, want one naming %v", tt.name, tt.vars)
			continue
		}
		for _, v := range tt.vars {
			if !strings.Contains(err.Error(), v) {
				t.Errorf("%s: error %q doesn't name %s: report every problem at once", tt.name, err, v)
			}
		}
	}
	_, err := LoadConfig(getenv(map[string]string{"TWENDE_API_KEY": "tooshort-but-secret"}))
	if err == nil || strings.Contains(err.Error(), "tooshort-but-secret") {
		t.Errorf("error %v must not contain the key itself", err)
	}
}

func TestConfigNeverLeaksTheKey(t *testing.T) {
	cfg, _ := LoadConfig(getenv(validEnv()))
	for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
		if out := fmt.Sprintf(format, cfg); strings.Contains(out, apiKey) {
			t.Errorf("fmt.Sprintf(%q, cfg) shows the API key", format)
		}
	}
	var text, js bytes.Buffer
	slog.New(slog.NewTextHandler(&text, nil)).Info("config", "config", cfg, "key", cfg.APIKey)
	slog.New(slog.NewJSONHandler(&js, nil)).Info("config", "config", cfg, "key", cfg.APIKey)
	for name, out := range map[string]string{"text": text.String(), "json": js.String()} {
		if strings.Contains(out, apiKey) {
			t.Errorf("logging the config with the %s handler shows the API key:\n%s", name, out)
		}
	}
	if b, _ := json.Marshal(cfg); strings.Contains(string(b), apiKey) {
		t.Errorf("json.Marshal(cfg) shows the API key")
	}
}

// --- Requirement 2: timeouts everywhere ---

func TestHTTPServerTimeouts(t *testing.T) {
	s, _, _ := newServer(t, nil, nil)
	srv := s.HTTPServer()
	checks := []struct {
		name      string
		got, want time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", srv.ReadTimeout, 10 * time.Second},
		{"WriteTimeout", srv.WriteTimeout, 10 * time.Second},
		{"IdleTimeout", srv.IdleTimeout, 60 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if srv.MaxHeaderBytes <= 0 || srv.MaxHeaderBytes > 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want a limit of at most 1 MiB", srv.MaxHeaderBytes)
	}
	if srv.Addr != "localhost:8080" || srv.Handler == nil {
		t.Errorf("Addr %q, Handler %v: want the config's address and the server's handler", srv.Addr, srv.Handler)
	}
}

// slowStore never answers until the caller gives up.
type slowStore struct{ *MemStore }

func (s slowStore) ListTrips(ctx context.Context) ([]Trip, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (s slowStore) Ping(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestHandlerTimeout(t *testing.T) {
	_, h, logs := newServer(t, map[string]string{"TWENDE_HANDLER_TIMEOUT": "30ms"}, slowStore{NewMemStore(trips...)})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- do(h, "GET", "/trips", "") }()
	select {
	case rec := <-done:
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("slow store: status %d, want 503", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "deadline") || strings.Contains(rec.Body.String(), "context") {
			t.Errorf("body %q leaks the internal error; log it, answer a plain message", rec.Body.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GET /trips with a hung store never answered: give store calls a context.WithTimeout of HandlerTimeout")
	}
	if !strings.Contains(logs.String(), "deadline exceeded") {
		t.Errorf("the store error wasn't logged; logs:\n%s", logs.String())
	}
}

// --- Requirement 3: rate limiting ---

func TestRateLimit(t *testing.T) {
	_, h, _ := newServer(t, map[string]string{"TWENDE_RATE": "0.01", "TWENDE_BURST": "2"}, nil)
	from := func(ip, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		req.RemoteAddr = ip + ":40000"
		req.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(len(path))) // must be ignored
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	for i := 1; i <= 2; i++ {
		if rec := from("198.51.100.7", "/trips"); rec.Code != 200 {
			t.Fatalf("request %d: status %d, want 200 (within the burst)", i, rec.Code)
		}
	}
	rec := from("198.51.100.7", "/trips")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status %d, want 429", rec.Code)
	}
	if n, err := strconv.Atoi(rec.Header().Get("Retry-After")); err != nil || n < 1 {
		t.Errorf("Retry-After = %q, want a whole number of seconds", rec.Header().Get("Retry-After"))
	}
	var body map[string]string
	if json.Unmarshal(rec.Body.Bytes(), &body) != nil || body["error"] == "" {
		t.Errorf("429 body = %q, want a JSON error", rec.Body.String())
	}
	if rec := from("198.51.100.8", "/trips"); rec.Code != 200 {
		t.Errorf("another IP: status %d, want 200: limit per client", rec.Code)
	}
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		if rec := from("198.51.100.7", path); rec.Code != 200 {
			t.Errorf("%s from a limited client: status %d, want 200: probes and scrapes are never limited", path, rec.Code)
		}
	}
}

// --- Requirement 4: security headers ---

func TestSecurityHeaders(t *testing.T) {
	_, h, _ := newServer(t, map[string]string{"TWENDE_BURST": "1", "TWENDE_RATE": "0.01"}, nil)
	want := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Content-Security-Policy":   "default-src 'none'; frame-ancestors 'none'",
		"Referrer-Policy":           "no-referrer",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Cache-Control":             "no-store",
	}
	responses := map[string]*httptest.ResponseRecorder{
		"200 /trips":    do(h, "GET", "/trips", ""),
		"429 /trips":    do(h, "GET", "/trips", ""),
		"404 /nowhere":  do(h, "GET", "/nowhere", ""),
		"200 /healthz":  do(h, "GET", "/healthz", ""),
		"401 /bookings": do(h, "POST", "/bookings", "{}"),
	}
	for name, rec := range responses {
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s: %s = %q, want %q", name, k, got, v)
			}
		}
	}
}

// --- Requirement 5: the booking endpoint checks its caller and its input ---

func TestBookingAuth(t *testing.T) {
	_, h, _ := newServer(t, nil, nil)
	body := `{"tripId":"46K-0730","seat":3,"phone":"0712345678"}`
	for _, auth := range []string{"", "Bearer wrong", "Bearer " + apiKey[:len(apiKey)-1], apiKey, "Basic " + apiKey} {
		rec := do(h, "POST", "/bookings", body, "Authorization", auth)
		if rec.Code != http.StatusUnauthorized || rec.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("Authorization %q: %d, WWW-Authenticate %q, want 401 and Bearer", auth, rec.Code, rec.Header().Get("WWW-Authenticate"))
		}
	}
	if rec := book(h, body); rec.Code != http.StatusCreated {
		t.Errorf("right key: %d %s, want 201", rec.Code, rec.Body.String())
	}
}

func TestBookingInput(t *testing.T) {
	_, h, _ := newServer(t, nil, nil)
	if rec := book(h, `{"tripId":"46K-0730","seat":3,"phone":"0712345678"}`); rec.Code != 201 || rec.Header().Get("Location") != "/bookings/B1" {
		t.Fatalf("first booking: %d, Location %q, want 201 and /bookings/B1", rec.Code, rec.Header().Get("Location"))
	}
	tests := []struct {
		name, body string
		code       int
	}{
		{"taken", `{"tripId":"46K-0730","seat":3,"phone":"0722000111"}`, 409},
		{"no trip", `{"tripId":"999","seat":3,"phone":"0722000111"}`, 404},
		{"seat 0", `{"tripId":"46K-0730","seat":0,"phone":"0722000111"}`, 422},
		{"seat 15", `{"tripId":"46K-0730","seat":15,"phone":"0722000111"}`, 422},
		{"bad phone", `{"tripId":"46K-0730","seat":4,"phone":"+254 712"}`, 422},
		{"no trip id", `{"seat":4,"phone":"0722000111"}`, 422},
		{"unknown field", `{"tripId":"46K-0730","seat":4,"phone":"0722000111","paid":true}`, 400},
		{"not JSON", `seat=4`, 400},
		{"01 numbers work", `{"tripId":"46K-0730","seat":4,"phone":"0110000111"}`, 201},
	}
	for _, tt := range tests {
		if rec := book(h, tt.body); rec.Code != tt.code {
			t.Errorf("%s: %d %s, want %d", tt.name, rec.Code, rec.Body.String(), tt.code)
		}
	}
}

func TestBodyLimit(t *testing.T) {
	_, h, _ := newServer(t, nil, nil)
	huge := `{"tripId":"46K-0730","seat":5,"phone":"0712345678","x":"` + strings.Repeat("a", 2<<20) + `"}`
	if rec := book(h, huge); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("a 2 MiB body: %d, want 413", rec.Code)
	}
}

// --- Requirement 6: one structured log line per request, nothing secret ---

func TestLogs(t *testing.T) {
	_, h, logs := newServer(t, nil, nil)
	rec := do(h, "POST", "/bookings?token=URL-SECRET", `{"tripId":"46K-0730","seat":3,"phone":"0712345678"}`, "Authorization", "Bearer "+apiKey)
	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("no X-Request-ID response header")
	}
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("logged %d lines for one request, want 1:\n%s", len(lines), logs.String())
	}
	var l map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &l); err != nil {
		t.Fatalf("log line isn't JSON: %s", lines[0])
	}
	for k, v := range map[string]any{"msg": "request", "request_id": id, "method": "POST", "path": "/bookings", "status": float64(201)} {
		if l[k] != v {
			t.Errorf("log %s = %v, want %v", k, l[k], v)
		}
	}
	if _, ok := l["duration_ms"]; !ok {
		t.Errorf("log has no duration_ms")
	}
	for _, secret := range []string{apiKey, "URL-SECRET", "0712345678"} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("the log contains %q", secret)
		}
	}
}

// panicStore panics, like a nil map or a bad index somewhere deep in a handler.
type panicStore struct{ *MemStore }

func (panicStore) ListTrips(context.Context) ([]Trip, error) {
	panic("index out of range [3] with length 3")
}

func TestPanicRecovery(t *testing.T) {
	_, h, logs := newServer(t, nil, panicStore{NewMemStore(trips...)})
	rec := do(h, "GET", "/trips", "")
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "index out of range") {
		t.Errorf("panic: %d %q, want 500 without the panic text", rec.Code, rec.Body.String())
	}
	if !strings.Contains(logs.String(), "index out of range") || !strings.Contains(logs.String(), rec.Header().Get("X-Request-ID")) {
		t.Errorf("the panic and its request ID weren't logged:\n%s", logs.String())
	}
	if rec := do(h, "GET", "/healthz", ""); rec.Code != 200 {
		t.Errorf("after a panic /healthz = %d, want 200", rec.Code)
	}
}

// --- Requirement 7: metrics ---

func TestMetrics(t *testing.T) {
	_, h, _ := newServer(t, nil, nil)
	do(h, "GET", "/trips", "")
	do(h, "GET", "/trips", "")
	book(h, `{"tripId":"999","seat":1,"phone":"0712345678"}`)
	rec := do(h, "GET", "/metrics", "")
	if rec.Code != 200 {
		t.Fatalf("/metrics = %d", rec.Code)
	}
	for _, line := range []string{
		`http_requests_total{code="200",method="GET",route="/trips"} 2`,
		`http_requests_total{code="404",method="POST",route="/bookings"} 1`,
	} {
		if !strings.Contains(rec.Body.String(), line+"\n") {
			t.Errorf("/metrics is missing\n\t%s", line)
		}
	}
}

// --- Requirement 8: health and readiness ---

type downStore struct{ *MemStore }

func (downStore) Ping(context.Context) error {
	return errors.New("dial tcp 10.0.3.7:5432: connection refused")
}

func TestHealthAndReadiness(t *testing.T) {
	_, h, _ := newServer(t, nil, nil)
	if rec := do(h, "GET", "/healthz", ""); rec.Code != 200 || rec.Body.String() != "ok\n" {
		t.Errorf("/healthz = %d %q, want 200 \"ok\\n\"", rec.Code, rec.Body.String())
	}
	if rec := do(h, "GET", "/readyz", ""); rec.Code != 200 {
		t.Errorf("/readyz = %d, want 200", rec.Code)
	}
	_, down, _ := newServer(t, nil, downStore{NewMemStore()})
	rec := do(down, "GET", "/readyz", "")
	if rec.Code != 503 || strings.Contains(rec.Body.String(), "10.0.3.7") {
		t.Errorf("/readyz with the database down = %d %q, want 503 without internal details", rec.Code, rec.Body.String())
	}
	if rec := do(down, "GET", "/healthz", ""); rec.Code != 200 {
		t.Errorf("/healthz with the database down = %d, want 200: liveness doesn't check dependencies", rec.Code)
	}
}

// --- Requirement 9: graceful shutdown ---

// gateStore holds ListTrips open until the test closes gate.
type gateStore struct {
	*MemStore
	entered chan struct{}
	gate    chan struct{}
}

func (g gateStore) ListTrips(ctx context.Context) ([]Trip, error) {
	g.entered <- struct{}{}
	<-g.gate
	return g.MemStore.ListTrips(ctx)
}

func TestGracefulShutdown(t *testing.T) {
	store := gateStore{NewMemStore(trips...), make(chan struct{}, 1), make(chan struct{})}
	s, h, _ := newServer(t, map[string]string{"TWENDE_HANDLER_TIMEOUT": "5s", "TWENDE_SHUTDOWN_GRACE": "5s"}, store)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	url := "http://" + ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() { runErr <- s.Run(ctx, ln) }()

	// Start a request and hold it inside the handler.
	client := &http.Client{Timeout: 5 * time.Second}
	type result struct {
		code int
		body string
		err  error
	}
	inflight := make(chan result, 1)
	go func() {
		resp, err := client.Get(url + "/trips")
		if err != nil {
			inflight <- result{err: err}
			return
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		inflight <- result{code: resp.StatusCode, body: string(b)}
	}()
	select {
	case <-store.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the request never reached the store")
	}

	cancel() // SIGTERM arrives

	// Readiness must fail while draining, so the load balancer stops sending traffic.
	deadline := time.Now().Add(2 * time.Second)
	for do(h, "GET", "/readyz", "").Code != http.StatusServiceUnavailable {
		if time.Now().After(deadline) {
			t.Fatal("/readyz still 200 after shutdown began: fail readiness first")
		}
		runtime.Gosched()
	}
	select {
	case err := <-runErr:
		t.Fatalf("Run returned %v while a request was still in flight: use Shutdown, not Close", err)
	default:
	}

	close(store.gate) // let the in-flight request finish
	r := <-inflight
	if r.err != nil || r.code != 200 || !strings.Contains(r.body, "46K-0730") {
		t.Errorf("in-flight request: %d %q %v, want 200 with the trips: let it finish", r.code, r.body, r.err)
	}
	select {
	case err := <-runErr:
		if err != nil {
			t.Errorf("Run = %v, want nil after a clean shutdown", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run didn't return after the last request finished")
	}
	if _, err := client.Get(url + "/healthz"); err == nil {
		t.Errorf("the server still accepts connections after Run returned")
	}
}
