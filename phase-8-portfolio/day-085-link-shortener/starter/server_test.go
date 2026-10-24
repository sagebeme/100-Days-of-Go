package linkshortener

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

var created = time.Date(2026, 10, 24, 9, 30, 0, 0, time.UTC)

// codes returns a NewCode func that hands out the given codes in order.
func codes(list ...string) func() string {
	var mu sync.Mutex
	i := 0
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		c := list[i%len(list)]
		i++
		return c
	}
}

// allowAll is a Limiter that never says no.
type allowAll struct{}

func (allowAll) Allow(string) (bool, time.Duration) { return true, 0 }

func newTestServer(t *testing.T, cfg Config) (*Server, Store) {
	t.Helper()
	if cfg.Store == nil {
		cfg.Store = NewMemoryStore()
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://kiungo.ke"
	}
	if cfg.Limiter == nil {
		cfg.Limiter = allowAll{}
	}
	if cfg.Lookup == nil {
		cfg.Lookup = dns
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return created }
	}
	cfg.Logf = func(string, ...any) {}
	return New(cfg), cfg.Store
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "196.201.214.9:51234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if err := s.Create(ctx, Link{Code: "abc", URL: "https://duka.co.ke"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(ctx, Link{Code: "abc", URL: "https://other.co.ke"}); !errors.Is(err, ErrCodeTaken) {
		t.Errorf("second Create with the same code = %v, want ErrCodeTaken", err)
	}
	if l, _ := s.Get(ctx, "abc"); l.URL != "https://duka.co.ke" {
		t.Errorf("Create overwrote the link: URL = %q", l.URL)
	}
	if _, err := s.Get(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(nope) = %v, want ErrNotFound", err)
	}
	if _, err := s.Hit(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Hit(nope) = %v, want ErrNotFound", err)
	}

	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { s.Hit(ctx, "abc") })
	}
	wg.Wait()
	if l, _ := s.Get(ctx, "abc"); l.Hits != 100 {
		t.Errorf("after 100 concurrent hits, Hits = %d, want 100", l.Hits)
	}
}

func TestShortenRetriesOnCollision(t *testing.T) {
	store := NewMemoryStore()
	store.Create(context.Background(), Link{Code: "taken01", URL: "https://duka.co.ke/old"})
	srv, _ := newTestServer(t, Config{Store: store, NewCode: codes("taken01", "taken01", "fresh01")})

	l, err := srv.Shorten(context.Background(), "https://duka.co.ke/new")
	if err != nil {
		t.Fatalf("Shorten: %v", err)
	}
	if l.Code != "fresh01" {
		t.Errorf("Code = %q, want fresh01 after two collisions", l.Code)
	}
	if old, _ := store.Get(context.Background(), "taken01"); old.URL != "https://duka.co.ke/old" {
		t.Errorf("a collision overwrote the old link: %q", old.URL)
	}
}

func TestShortenGivesUp(t *testing.T) {
	store := NewMemoryStore()
	store.Create(context.Background(), Link{Code: "taken01", URL: "https://duka.co.ke/old"})
	srv, _ := newTestServer(t, Config{Store: store, NewCode: codes("taken01"), MaxTries: 3})

	if _, err := srv.Shorten(context.Background(), "https://duka.co.ke/new"); !errors.Is(err, ErrNoFreeCode) {
		t.Errorf("Shorten with every code taken = %v, want ErrNoFreeCode", err)
	}
}

func TestCreateRedirectStats(t *testing.T) {
	srv, _ := newTestServer(t, Config{NewCode: codes("aB3xY9q")})

	rec := do(srv, "POST", "/links", `{"url":"https://duka.co.ke/sale?item=sufuria"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /links status = %d, want 201; body %s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Location"); got != "https://kiungo.ke/aB3xY9q" {
		t.Errorf("Location = %q, want https://kiungo.ke/aB3xY9q", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body struct{ Code, Short, URL string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response isn't JSON: %v", err)
	}
	want := struct{ Code, Short, URL string }{"aB3xY9q", "https://kiungo.ke/aB3xY9q", "https://duka.co.ke/sale?item=sufuria"}
	if body != want {
		t.Errorf("body = %+v, want %+v", body, want)
	}

	for range 3 {
		rec = do(srv, "GET", "/aB3xY9q", "")
		if rec.Code != http.StatusFound {
			t.Fatalf("GET /aB3xY9q status = %d, want 302", rec.Code)
		}
		if got := rec.Header().Get("Location"); got != "https://duka.co.ke/sale?item=sufuria" {
			t.Errorf("redirect Location = %q", got)
		}
	}

	rec = do(srv, "GET", "/links/aB3xY9q/stats", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("stats status = %d, want 200", rec.Code)
	}
	var stats Link
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatalf("stats isn't JSON: %v", err)
	}
	if stats.Hits != 3 || stats.Code != "aB3xY9q" || !stats.Created.Equal(created) {
		t.Errorf("stats = %+v, want 3 hits, code aB3xY9q, created %v", stats, created)
	}
}

func TestErrors(t *testing.T) {
	srv, _ := newTestServer(t, Config{NewCode: codes("aB3xY9q")})
	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"unknown code", "GET", "/zzzzzzz", "", http.StatusNotFound},
		{"code with a dot", "GET", "/favicon.ico", "", http.StatusNotFound},
		{"stats for unknown code", "GET", "/links/zzzzzzz/stats", "", http.StatusNotFound},
		{"broken JSON", "POST", "/links", `{"url":`, http.StatusBadRequest},
		{"javascript URL", "POST", "/links", `{"url":"javascript:alert(1)"}`, http.StatusUnprocessableEntity},
		{"metadata address", "POST", "/links", `{"url":"http://169.254.169.254/"}`, http.StatusUnprocessableEntity},
		{"private name", "POST", "/links", `{"url":"https://internal.example/"}`, http.StatusUnprocessableEntity},
		{"huge body", "POST", "/links", `{"url":"https://duka.co.ke/` + strings.Repeat("a", MaxBodyBytes) + `"}`, http.StatusRequestEntityTooLarge},
		{"wrong method", "DELETE", "/links", "", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(srv, tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Errorf("%s %s status = %d, want %d; body %s", tt.method, tt.path, rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	// 3 at once, then one an hour: the 4th request in this test is refused.
	srv, _ := newTestServer(t, Config{Limiter: NewIPLimiter(1.0/3600, 3), NewCode: RandomCode})
	for i := 1; i <= 3; i++ {
		if rec := do(srv, "POST", "/links", `{"url":"https://duka.co.ke"}`); rec.Code != http.StatusCreated {
			t.Fatalf("request %d status = %d, want 201", i, rec.Code)
		}
	}
	rec := do(srv, "POST", "/links", `{"url":"https://duka.co.ke"}`)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request status = %d, want 429", rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" || strings.Contains(ra, ".") {
		t.Errorf("Retry-After = %q, want whole seconds, at least 1", ra)
	}

	// A different client has its own bucket.
	req := httptest.NewRequest("POST", "/links", strings.NewReader(`{"url":"https://duka.co.ke"}`))
	req.RemoteAddr = "41.90.1.1:4000"
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("another client's first request status = %d, want 201", rec.Code)
	}
}
