package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// logLines decodes every JSON log line in buf.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line %q isn't JSON: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

func newLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func TestChainOrder(t *testing.T) {
	var got []string
	mark := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, name+" in")
				next.ServeHTTP(w, r)
				got = append(got, name+" out")
			})
		}
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = append(got, "handler") })

	Chain(h, mark("a"), mark("b"), mark("c")).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	want := []string{"a in", "b in", "c in", "handler", "c out", "b out", "a out"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chain(h, a, b, c) ran\n%q\nwant\n%q", got, want)
	}

	got = nil
	Chain(h).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !reflect.DeepEqual(got, []string{"handler"}) {
		t.Errorf("Chain(h) with no middleware ran %q, want just the handler", got)
	}
}

func isHex32(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func TestRequestID(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	tests := []struct {
		name     string
		incoming string
		keep     bool
	}{
		{"none sent", "", false},
		{"a good one", "req-2026-11-02_07.30-abc", true},
		{"64 characters", strings.Repeat("a", 64), true},
		{"65 characters", strings.Repeat("a", 65), false},
		{"has a space", "abc def", false},
		{"tries to forge a log line", "abc\nlevel=ERROR msg=hacked", false},
		{"quotes", `abc"def`, false},
	}
	ids := map[string]bool{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen = ""
			req := httptest.NewRequest("GET", "/", nil)
			if tt.incoming != "" {
				req.Header.Set("X-Request-ID", tt.incoming)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			header := rec.Header().Get("X-Request-ID")
			if header != seen {
				t.Errorf("response header %q and RequestIDFrom %q differ", header, seen)
			}
			if tt.keep {
				if seen != tt.incoming {
					t.Errorf("ID = %q, want the incoming %q kept", seen, tt.incoming)
				}
				return
			}
			if !isHex32(seen) {
				t.Errorf("ID = %q, want a new one of 32 lower-case hex digits", seen)
			}
			if ids[seen] {
				t.Errorf("ID %q was handed out twice", seen)
			}
			ids[seen] = true
		})
	}

	if got := RequestIDFrom(context.Background()); got != "" {
		t.Errorf("RequestIDFrom(a context without an ID) = %q, want \"\"", got)
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		status  float64
		bytes   float64
		level   string
	}{
		{"created", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, "hello")
		}, 201, 5, "INFO"},
		{"writes without WriteHeader", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "ok\n")
		}, 200, 3, "INFO"},
		{"writes nothing at all", func(w http.ResponseWriter, r *http.Request) {}, 200, 0, "INFO"},
		{"not found", http.NotFound, 404, 19, "INFO"},
		{"server error", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}, 503, 0, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := newLogger()
			h := Chain(tt.handler, RequestID, Logger(logger))
			req := httptest.NewRequest("POST", "/trips/1/bookings?token=sir1-ya-mtu", nil)
			req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.secret")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			lines := logLines(t, buf)
			if len(lines) != 1 {
				t.Fatalf("got %d log lines, want 1:\n%s", len(lines), buf)
			}
			l := lines[0]
			checks := []struct {
				key  string
				want any
			}{
				{"msg", "request"},
				{"level", tt.level},
				{"method", "POST"},
				{"path", "/trips/1/bookings"},
				{"status", tt.status},
				{"bytes", tt.bytes},
				{"request_id", rec.Header().Get("X-Request-ID")},
			}
			for _, c := range checks {
				if l[c.key] != c.want {
					t.Errorf("log %s = %v, want %v", c.key, l[c.key], c.want)
				}
			}
			if _, ok := l["duration"]; !ok {
				t.Errorf("log line has no duration: %s", buf)
			}
			if rec.Code != int(tt.status) {
				t.Errorf("the client got %d, want %d: the logger must pass the status through", rec.Code, int(tt.status))
			}
			for _, secret := range []string{"sir1-ya-mtu", "eyJhbGci", "secret"} {
				if strings.Contains(buf.String(), secret) {
					t.Errorf("the log leaks %q: %s", secret, buf)
				}
			}
		})
	}
}

func TestRecover(t *testing.T) {
	logger, buf := newLogger()
	boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var seats map[string]int
		seats["3"] = 1 // assignment to entry in nil map
	})
	h := Chain(boom, RequestID, Logger(logger), Recover(logger))
	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("the panic got past Recover: %v", p)
			}
		}()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/trips/1", nil))
	}()

	if rec.Code != 500 {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"something went wrong"}` {
		t.Errorf("body = %q, want {\"error\":\"something went wrong\"}", got)
	}
	if strings.Contains(rec.Body.String(), "nil map") {
		t.Error("the panic message reached the client")
	}

	lines := logLines(t, buf)
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2 (the panic, then the request):\n%s", len(lines), buf)
	}
	p, r := lines[0], lines[1]
	id := rec.Header().Get("X-Request-ID")
	if p["msg"] != "panic" || p["level"] != "ERROR" || p["request_id"] != id {
		t.Errorf("first log line = %v, want msg panic, level ERROR, request_id %s", p, id)
	}
	if v, _ := p["value"].(string); !strings.Contains(v, "nil map") {
		t.Errorf("panic log value = %v, want the panic message", p["value"])
	}
	if s, _ := p["stack"].(string); !strings.Contains(s, "goroutine") {
		t.Errorf("panic log has no stack trace")
	}
	if r["msg"] != "request" || r["status"] != 500.0 {
		t.Errorf("second log line = %v, want the request logged with status 500", r)
	}
}

func TestRecoverPassesAbortHandlerOn(t *testing.T) {
	logger, _ := newLogger()
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if p := recover(); p != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler passed on", p)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}

func TestMaskPhone(t *testing.T) {
	tests := []struct{ in, want string }{
		{"254712345678", "2547*****678"},
		{"0712345678", "0712***678"},
		{"12345678", "1234*678"},
		{"0712", "****"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := MaskPhone(tt.in); got != tt.want {
			t.Errorf("MaskPhone(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	logger, buf := newLogger()
	h := New(logger)
	req := httptest.NewRequest("POST", "/bookings", strings.NewReader(`{"passenger": "Achieng", "phone": "254712345678"}`))
	req.Header.Set("X-Request-ID", "trace-42")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("POST /bookings: status = %d, want 201", rec.Code)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "trace-42" {
		t.Errorf("X-Request-ID = %q, want trace-42: is RequestID in the chain?", got)
	}
	if strings.Contains(buf.String(), "254712345678") {
		t.Errorf("the log holds the full phone number: %s", buf)
	}
	lines := logLines(t, buf)
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2 (booking created, request):\n%s", len(lines), buf)
	}
	if lines[0]["phone"] != "2547*****678" || lines[0]["request_id"] != "trace-42" {
		t.Errorf("booking log = %v, want phone 2547*****678 and request_id trace-42", lines[0])
	}
	if lines[1]["msg"] != "request" || lines[1]["request_id"] != "trace-42" || lines[1]["status"] != 201.0 {
		t.Errorf("request log = %v, want msg request, request_id trace-42, status 201", lines[1])
	}
}
