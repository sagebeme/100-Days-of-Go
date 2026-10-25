package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ok(w http.ResponseWriter, r *http.Request) { w.Write([]byte("sawa")) }

func TestRequestID(t *testing.T) {
	var seen string
	h := RequestID(func() string { return "generated-1" }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	tests := []struct {
		name, header, want string
	}{
		{"none sent", "", "generated-1"},
		{"valid one kept", "abc-123-XYZ", "abc-123-XYZ"},
		{"newline rejected", "abc\nlevel=ERROR msg=fake", "generated-1"},
		{"space rejected", "abc 123", "generated-1"},
		{"too long rejected", strings.Repeat("a", 65), "generated-1"},
		{"64 is fine", strings.Repeat("a", 64), strings.Repeat("a", 64)},
	}
	for _, tt := range tests {
		seen = ""
		req := httptest.NewRequest("GET", "/trips", nil)
		if tt.header != "" {
			req.Header.Set("X-Request-ID", tt.header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if seen != tt.want {
			t.Errorf("%s: RequestIDFrom = %q, want %q", tt.name, seen, tt.want)
		}
		if got := rec.Header().Get("X-Request-ID"); got != tt.want {
			t.Errorf("%s: response X-Request-ID = %q, want %q", tt.name, got, tt.want)
		}
	}
	if got := RequestIDFrom(context.Background()); got != "" {
		t.Errorf("RequestIDFrom(empty context) = %q, want \"\"", got)
	}
	if a, b := RandomID(), RandomID(); len(a) != 32 || a == b {
		t.Errorf("RandomID() gave %q and %q, want two different 32-character IDs", a, b)
	}
}

// logLines runs one request through RequestID and Logging and returns the JSON lines logged.
func logLines(t *testing.T, h http.Handler, req *http.Request) []map[string]any {
	t.Helper()
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	RequestID(func() string { return "req-42" }, Logging(log, h)).ServeHTTP(httptest.NewRecorder(), req)
	var lines []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("log line isn't JSON: %q", l)
		}
		lines = append(lines, m)
	}
	return lines
}

func TestLogging(t *testing.T) {
	req := httptest.NewRequest("GET", "/trips?token=SECRET-TOKEN", nil)
	req.Header.Set("Authorization", "Bearer SECRET-JWT")
	lines := logLines(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}), req)
	if len(lines) != 1 {
		t.Fatalf("logged %d lines, want 1 per request", len(lines))
	}
	l := lines[0]
	want := map[string]any{"msg": "request", "level": "INFO", "request_id": "req-42", "method": "GET", "path": "/trips", "status": float64(404)}
	for k, v := range want {
		if l[k] != v {
			t.Errorf("log %s = %v, want %v", k, l[k], v)
		}
	}
	if _, ok := l["duration_ms"]; !ok {
		t.Errorf("log has no duration_ms: %v", l)
	}
	all, _ := json.Marshal(l)
	for _, secret := range []string{"SECRET-TOKEN", "SECRET-JWT"} {
		if strings.Contains(string(all), secret) {
			t.Errorf("the log contains %q: never log query strings or headers", secret)
		}
	}
}

func TestLoggingStatusAndLevel(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		status  float64
		level   string
	}{
		{"write without WriteHeader", ok, 200, "INFO"},
		{"nothing written", func(w http.ResponseWriter, r *http.Request) {}, 200, "INFO"},
		{"500", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "boom", 500) }, 500, "ERROR"},
		{"503", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }, 503, "ERROR"},
		{"201", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) }, 201, "INFO"},
	}
	for _, tt := range tests {
		lines := logLines(t, tt.handler, httptest.NewRequest("POST", "/bookings", nil))
		if len(lines) != 1 {
			t.Errorf("%s: logged %d lines, want 1", tt.name, len(lines))
			continue
		}
		if lines[0]["status"] != tt.status || lines[0]["level"] != tt.level {
			t.Errorf("%s: status %v level %v, want %v %v", tt.name, lines[0]["status"], lines[0]["level"], tt.status, tt.level)
		}
	}
}

// stepClock returns a clock that moves forward by each step in turn, one per call.
func stepClock(steps ...time.Duration) func() time.Time {
	now := time.Date(2026, 10, 24, 17, 0, 0, 0, time.UTC)
	i := 0
	return func() time.Time {
		if i > 0 && i-1 < len(steps) {
			now = now.Add(steps[i-1])
		}
		i++
		return now
	}
}

func TestInstrument(t *testing.T) {
	reg := prometheus.NewRegistry()
	// Three requests: 20ms, 300ms, 2s. Each request reads the clock twice.
	m := NewMetrics(reg, stepClock(20*time.Millisecond, 0, 300*time.Millisecond, 0, 2*time.Second))
	trips := m.Instrument("/trips/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/trips/999" {
			http.Error(w, "no such trip", http.StatusNotFound)
			return
		}
		ok(w, r)
	}))
	for _, path := range []string{"/trips/1", "/trips/2", "/trips/999"} {
		trips.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}

	want := `
# HELP http_requests_total HTTP requests by method, route and status code.
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET",route="/trips/{id}"} 2
http_requests_total{code="404",method="GET",route="/trips/{id}"} 1
# HELP http_request_duration_seconds How long HTTP requests took, by method and route.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="0.05"} 1
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="0.1"} 1
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="0.25"} 1
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="0.5"} 2
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="1"} 2
http_request_duration_seconds_bucket{method="GET",route="/trips/{id}",le="+Inf"} 3
http_request_duration_seconds_sum{method="GET",route="/trips/{id}"} 2.32
http_request_duration_seconds_count{method="GET",route="/trips/{id}"} 3
`
	got := scrape(t, reg)
	for _, line := range strings.Split(strings.TrimSpace(want), "\n") {
		if !strings.Contains(got, line+"\n") {
			t.Errorf("/metrics is missing the line\n\t%s\ngot:\n%s", line, got)
		}
	}
	if n := strings.Count(got, "http_requests_total{"); n != 2 {
		t.Errorf("%d request series, want 2: label by route pattern, never by the real path", n)
	}
}

// scrape returns what Prometheus would read from /metrics for this registry.
func scrape(t *testing.T, reg *prometheus.Registry) string {
	t.Helper()
	rec := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	return rec.Body.String()
}

func TestLive(t *testing.T) {
	rec := httptest.NewRecorder()
	Live(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 || rec.Body.String() != "ok\n" {
		t.Errorf("Live = %d %q, want 200 \"ok\\n\"", rec.Code, rec.Body.String())
	}
}

func readyz(t *testing.T, rd *Readiness) (int, readyBody) {
	t.Helper()
	rec := httptest.NewRecorder()
	rd.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	var body readyBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("readyz body %q isn't JSON: %v", rec.Body.String(), err)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	return rec.Code, body
}

func TestReadiness(t *testing.T) {
	var logs bytes.Buffer
	rd := NewReadiness(slog.New(slog.NewTextHandler(&logs, nil)), 50*time.Millisecond)
	rd.Add("db", func(ctx context.Context) error { return nil })
	rd.Add("redis", func(ctx context.Context) error { return nil })

	code, body := readyz(t, rd)
	if code != 200 || body.Status != "ok" || body.Checks["db"] != "ok" || body.Checks["redis"] != "ok" {
		t.Errorf("all healthy: %d %+v, want 200 ok with both checks ok", code, body)
	}

	rd.Add("redis", func(ctx context.Context) error { return errors.New("dial tcp 10.0.3.7:6379: connection refused") })
	code, body = readyz(t, rd)
	if code != 503 || body.Status != "fail" || body.Checks["redis"] != "fail" || body.Checks["db"] != "ok" {
		t.Errorf("redis down: %d %+v, want 503 fail, redis fail, db ok", code, body)
	}
	if !strings.Contains(logs.String(), "10.0.3.7") {
		t.Errorf("the failing check's error wasn't logged; logs: %q", logs.String())
	}

	// A check that hangs must be cut off by the timeout, not hang the probe.
	rd.Add("redis", func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
	done := make(chan int)
	go func() { c, _ := readyz(t, rd); done <- c }()
	select {
	case c := <-done:
		if c != 503 {
			t.Errorf("hanging check: %d, want 503", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a hanging check hung the readiness probe: give each check a context with a timeout")
	}
}

func TestReadinessLeaksNothing(t *testing.T) {
	rd := NewReadiness(slog.New(slog.NewTextHandler(io.Discard, nil)), time.Second)
	rd.Add("db", func(ctx context.Context) error {
		return errors.New("pq: password authentication failed for user twende_admin")
	})
	rec := httptest.NewRecorder()
	rd.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	if strings.Contains(rec.Body.String(), "twende_admin") || strings.Contains(rec.Body.String(), "pq:") {
		t.Errorf("readyz body %q leaks the error text: log it, answer only \"fail\"", rec.Body.String())
	}
}

func TestDrain(t *testing.T) {
	rd := NewReadiness(slog.New(slog.NewTextHandler(io.Discard, nil)), time.Second)
	rd.Add("db", func(ctx context.Context) error { return nil })
	rd.Drain()
	code, body := readyz(t, rd)
	if code != 503 || body.Status != "fail" {
		t.Errorf("after Drain: %d %+v, want 503 fail so the load balancer stops sending traffic", code, body)
	}
}
