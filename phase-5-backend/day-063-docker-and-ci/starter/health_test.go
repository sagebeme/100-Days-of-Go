package dockerandci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serve(h *Health, path string) (int, ReadyResponse, *httptest.ResponseRecorder) {
	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	var body ReadyResponse
	json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body, rec
}

func ok(context.Context) error { return nil }

func TestNewCheck(t *testing.T) {
	c := NewCheck("db", func(context.Context) error { return errors.New("down") })
	if c.Name() != "db" || c.Check(t.Context()) == nil {
		t.Errorf("NewCheck: Name() = %q, Check error = %v; want \"db\" and the function's error", c.Name(), c.Check(t.Context()))
	}
}

func TestLiveness(t *testing.T) {
	failing := NewCheck("db", func(context.Context) error { return errors.New("connection refused") })
	h := &Health{Checks: []Checker{failing}, Timeout: time.Second, Log: slog.New(slog.DiscardHandler)}
	code, body, rec := serve(h, "/healthz")
	if code != 200 || body.Status != "ok" {
		t.Errorf("GET /healthz = %d %s, want 200 {\"status\":\"ok\"} even with the database down", code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
	}
}

func TestReadiness(t *testing.T) {
	tests := []struct {
		name       string
		checks     []Checker
		wantCode   int
		wantStatus string
		wantChecks map[string]string
	}{
		{"no dependencies", nil, 200, "ready", nil},
		{"all ok", []Checker{NewCheck("db", ok), NewCheck("mpesa", ok)}, 200, "ready", map[string]string{"db": "ok", "mpesa": "ok"}},
		{"db down", []Checker{
			NewCheck("db", func(context.Context) error {
				return errors.New("dial tcp 10.0.0.5:5432: password authentication failed for user gopher")
			}),
			NewCheck("mpesa", ok),
		}, 503, "not ready", map[string]string{"db": "failed", "mpesa": "ok"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			h := &Health{Checks: tt.checks, Timeout: time.Second, Log: slog.New(slog.NewTextHandler(&logs, nil))}
			code, body, rec := serve(h, "/readyz")
			if code != tt.wantCode || body.Status != tt.wantStatus {
				t.Fatalf("GET /readyz = %d %s, want %d status %q", code, rec.Body.String(), tt.wantCode, tt.wantStatus)
			}
			if len(tt.wantChecks) > 0 && !equal(body.Checks, tt.wantChecks) {
				t.Errorf("checks = %v, want %v", body.Checks, tt.wantChecks)
			}
			if strings.Contains(rec.Body.String(), "password") || strings.Contains(rec.Body.String(), "10.0.0.5") {
				t.Errorf("the response leaks the error: %s", rec.Body.String())
			}
			if tt.wantCode == 503 && !strings.Contains(logs.String(), "password authentication failed") {
				t.Errorf("the failed check's error should be logged, got log:\n%s", logs.String())
			}
			if rec.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store: a cached \"ready\" is a lie", rec.Header().Get("Cache-Control"))
			}
		})
	}
}

func equal(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestReadinessTimeout(t *testing.T) {
	hang := NewCheck("db", func(ctx context.Context) error {
		<-ctx.Done() // a database that never answers
		return ctx.Err()
	})
	h := &Health{Checks: []Checker{hang}, Timeout: 50 * time.Millisecond, Log: slog.New(slog.DiscardHandler)}
	done := make(chan int, 1)
	go func() {
		code, _, _ := serve(h, "/readyz")
		done <- code
	}()
	select {
	case code := <-done:
		if code != 503 {
			t.Errorf("a check that hangs: status %d, want 503", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GET /readyz is still waiting after 2s: give the checks a context with Timeout")
	}
}

func TestReadinessRunsChecksAtTheSameTime(t *testing.T) {
	// Each check waits until the other has started. Run one after the other, the
	// first waits forever (well, until the timeout) and readiness fails.
	aStarted, bStarted := make(chan struct{}), make(chan struct{})
	a := NewCheck("a", func(ctx context.Context) error {
		close(aStarted)
		select {
		case <-bStarted:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	b := NewCheck("b", func(ctx context.Context) error {
		close(bStarted)
		select {
		case <-aStarted:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	h := &Health{Checks: []Checker{a, b}, Timeout: time.Second, Log: slog.New(slog.DiscardHandler)}
	if code, body, _ := serve(h, "/readyz"); code != 200 {
		t.Errorf("status %d %v, want 200: run the checks concurrently, one goroutine each", code, body)
	}
}

func TestDraining(t *testing.T) {
	called := false
	c := NewCheck("db", func(context.Context) error { called = true; return nil })
	h := &Health{Checks: []Checker{c}, Timeout: time.Second, Log: slog.New(slog.DiscardHandler)}
	if code, _, _ := serve(h, "/readyz"); code != 200 {
		t.Fatalf("before draining: %d, want 200", code)
	}
	called = false
	h.SetDraining()
	code, body, _ := serve(h, "/readyz")
	if code != 503 || body.Status != "draining" {
		t.Errorf("after SetDraining: %d %q, want 503 \"draining\"", code, body.Status)
	}
	if called {
		t.Errorf("checks ran while draining; there's no need")
	}
	if code, _, _ := serve(h, "/healthz"); code != 200 {
		t.Errorf("/healthz while draining = %d, want 200: the process is still alive", code)
	}
}
