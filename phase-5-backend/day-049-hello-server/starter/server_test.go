package helloserver

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		appEnv  string
		want    Config
		wantErr []string // pieces the error must contain; nil means no error
	}{
		{"defaults", "", "", Config{Port: 8080, AppEnv: "development"}, nil},
		{"both set", "9000", "production", Config{Port: 9000, AppEnv: "production"}, nil},
		{"staging", "3000", "staging", Config{Port: 3000, AppEnv: "staging"}, nil},
		{"top port", "65535", "", Config{Port: 65535, AppEnv: "development"}, nil},
		{"port not a number", "abc", "", Config{}, []string{"PORT", `"abc"`}},
		{"port zero", "0", "", Config{}, []string{"PORT"}},
		{"port too big", "70000", "", Config{}, []string{"PORT"}},
		{"unknown env", "", "prod", Config{}, []string{"APP_ENV", `"prod"`}},
		{"both wrong", "8k", "live", Config{}, []string{"PORT", "APP_ENV"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)
			t.Setenv("APP_ENV", tt.appEnv)
			got, err := LoadConfig()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("LoadConfig() error = %v, want none", err)
				}
				if got != tt.want {
					t.Errorf("LoadConfig()\n got: %+v\nwant: %+v", got, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("LoadConfig() = %+v with no error, want an error mentioning %v", got, tt.wantErr)
			}
			for _, piece := range tt.wantErr {
				if !strings.Contains(err.Error(), piece) {
					t.Errorf("LoadConfig() error %q doesn't mention %s", err, piece)
				}
			}
			if got != (Config{}) {
				t.Errorf("LoadConfig() returned %+v with its error, want Config{}", got)
			}
		})
	}
}

func TestAddr(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{8080, ":8080"},
		{3000, ":3000"},
	}
	for _, tt := range tests {
		if got := (Config{Port: tt.port}).Addr(); got != tt.want {
			t.Errorf("Config{Port: %d}.Addr() = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		method string
		path   string
		status int
		body   string // "" means don't check the body
	}{
		{"GET", "/", 200, "Karibu! The matatu booking API is running.\n"},
		{"GET", "/health", 200, "ok\n"},
		{"GET", "/hello/Wanjiku", 200, "Habari, Wanjiku!\n"},
		{"GET", "/hello/Otieno", 200, "Habari, Otieno!\n"},
		{"HEAD", "/health", 200, ""},
		{"GET", "/nope", 404, ""},
		{"GET", "/health/extra", 404, ""},
		{"GET", "/hello/", 404, ""},
		{"POST", "/health", 405, ""},
	}
	h := Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if tt.body != "" && rec.Body.String() != tt.body {
				t.Errorf("body\n got: %q\nwant: %q", rec.Body.String(), tt.body)
			}
			if tt.status == 200 {
				if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
					t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", ct)
				}
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	h := http.NotFoundHandler()
	srv := NewServer(":8080", h)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", srv.Addr, ":8080")
	}
	if srv.Handler == nil {
		t.Error("Handler is nil, want the handler you were given")
	}
	timeouts := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", srv.ReadTimeout, 10 * time.Second},
		{"WriteTimeout", srv.WriteTimeout, 10 * time.Second},
		{"IdleTimeout", srv.IdleTimeout, 60 * time.Second},
	}
	for _, tt := range timeouts {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// listen opens a listener on a free port on this machine only.
func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

var client = &http.Client{Timeout: time.Second}

func TestRunServesThenStops(t *testing.T) {
	ln := listen(t)
	srv := &http.Server{Handler: Routes()}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, time.Second) }()

	resp, err := client.Get("http://" + ln.Addr().String() + "/health")
	if err != nil {
		cancel()
		t.Fatalf("GET /health on the running server: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(body) != "ok\n" {
		t.Errorf("GET /health = %d %q, want 200 %q", resp.StatusCode, body, "ok\n")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v after a clean shutdown, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run didn't return after its context was cancelled")
	}
}

func TestRunWaitsForInFlightRequests(t *testing.T) {
	ln := listen(t)
	started := make(chan struct{})
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		io.WriteString(w, "done\n")
	})
	srv := &http.Server{Handler: mux}
	shuttingDown := make(chan struct{})
	srv.RegisterOnShutdown(func() { close(shuttingDown) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, 2*time.Second) }()

	type result struct {
		status int
		body   string
		err    error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := client.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			got <- result{err: err}
			return
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		got <- result{status: resp.StatusCode, body: string(b)}
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("the request never reached the handler: is Run serving on the listener?")
	}

	cancel() // like pressing Ctrl+C while a booking is half done
	select {
	case <-shuttingDown:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Run didn't call srv.Shutdown after its context was cancelled")
	}
	select {
	case err := <-done:
		close(release)
		t.Fatalf("Run returned (%v) while a request was still running; Shutdown should wait for it", err)
	default:
	}

	close(release)
	r := <-got
	if r.err != nil {
		t.Fatalf("the in-flight request failed: %v", r.err)
	}
	if r.status != 200 || r.body != "done\n" {
		t.Errorf("in-flight request = %d %q, want 200 %q", r.status, r.body, "done\n")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run didn't return after the last request finished")
	}
}

func TestRunGivesUpAfterTimeout(t *testing.T) {
	ln := listen(t)
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stuck", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	})
	srv := &http.Server{Handler: mux}
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, 50*time.Millisecond) }()

	go func() {
		resp, err := client.Get("http://" + ln.Addr().String() + "/stuck")
		if err == nil {
			resp.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("the request never reached the handler: is Run serving on the listener?")
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Run returned %v, want an error wrapping context.DeadlineExceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run didn't give up after its shutdown timeout")
	}
}

func TestRunReportsServeErrors(t *testing.T) {
	ln := listen(t)
	ln.Close() // Serve fails at once on a closed listener
	srv := &http.Server{Handler: Routes()}

	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), srv, ln, time.Second) }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("Run returned nil on a closed listener, want the error from Serve")
		}
	case <-time.After(time.Second):
		t.Fatal("Run didn't return when Serve failed")
	}
}
