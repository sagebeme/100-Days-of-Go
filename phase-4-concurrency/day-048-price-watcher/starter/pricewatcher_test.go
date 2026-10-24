package pricewatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Every shop in these tests is a fake: an httptest.Server on 127.0.0.1, inside the test
// process. Nothing goes to the internet.

// serve starts a fake shop that answers every request with h, and stops it when the
// test ends.
func serve(t *testing.T, h http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv.URL
}

// prices returns a handler that serves body as JSON on GET /prices.
func prices(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/prices" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}
}

// noSleep records the backoff waits instead of waiting.
type noSleep struct {
	mu    sync.Mutex
	waits []time.Duration
}

func (s *noSleep) sleep(ctx context.Context, d time.Duration) error {
	s.mu.Lock()
	s.waits = append(s.waits, d)
	s.mu.Unlock()
	return ctx.Err()
}

func quick() Config {
	return Config{Workers: 4, Timeout: 2 * time.Second, Sleep: (&noSleep{}).sleep}
}

func TestWatchCheapest(t *testing.T) {
	shops := []Shop{
		{"Mama Njeri Shop", serve(t, prices(`[{"item":"unga 2kg","price":230},{"item":"sukari 1kg","price":190}]`))},
		{"Duka la Kona", serve(t, prices(`[{"item":"unga 2kg","price":245},{"item":"sukari 1kg","price":180},{"item":"chumvi 500g","price":30}]`))},
		{"Kariakor Traders", serve(t, prices(`[{"item":"unga 2kg","price":230},{"item":"mafuta 1L","price":310}]`))},
		{"Ngara Stores", serve(t, prices(`[{"item":"","price":5},{"item":"chumvi 500g","price":0},{"item":"mafuta 1L","price":-20}]`))},
		{"Empty Shelf", serve(t, prices(`[]`))},
	}
	got, err := Watch(context.Background(), http.DefaultClient, shops, quick())
	if err != nil {
		t.Fatalf("Watch error = %v, want nil", err)
	}
	want := Summary{Cheapest: map[string]Offer{
		"unga 2kg":    {"Kariakor Traders", 230}, // a tie with Mama Njeri: Kariakor sorts first
		"sukari 1kg":  {"Duka la Kona", 180},
		"chumvi 500g": {"Duka la Kona", 30},      // Ngara's 0 doesn't count
		"mafuta 1L":   {"Kariakor Traders", 310}, // nor does its -20
	}}
	if !reflect.DeepEqual(got.Cheapest, want.Cheapest) {
		t.Errorf("Cheapest\n got: %v\nwant: %v", got.Cheapest, want.Cheapest)
	}
	if len(got.Failed) != 0 {
		t.Errorf("Failed = %q, want none", got.Failed)
	}
}

func TestWatchNoShops(t *testing.T) {
	got, err := Watch(context.Background(), http.DefaultClient, nil, quick())
	if err != nil || len(got.Cheapest) != 0 || len(got.Failed) != 0 {
		t.Errorf("Watch(no shops) = %v, %v; want an empty summary and nil", got, err)
	}
}

func TestWatchFailures(t *testing.T) {
	good := serve(t, prices(`[{"item":"unga 2kg","price":230}]`))
	shops := []Shop{
		{"Zawadi Mart", good},
		{"Broken JSON", serve(t, prices(`[{"item":"unga 2kg","price":`))},
		{"Moved Away", serve(t, func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })},
		{"Closed Shop", "http://127.0.0.1:1"}, // nothing listens on port 1
	}
	got, err := Watch(context.Background(), http.DefaultClient, shops, quick())
	if err != nil {
		t.Fatalf("Watch error = %v, want nil: one shop failing isn't Watch failing", err)
	}
	if want := []string{"Broken JSON", "Closed Shop", "Moved Away"}; !slices.Equal(got.Failed, want) {
		t.Errorf("Failed = %q, want %q (sorted)", got.Failed, want)
	}
	if want := map[string]Offer{"unga 2kg": {"Zawadi Mart", 230}}; !reflect.DeepEqual(got.Cheapest, want) {
		t.Errorf("Cheapest = %v, want %v", got.Cheapest, want)
	}
}

func TestWatchBounded(t *testing.T) {
	// 12 shops, 3 workers. Every handler holds its request until the test lets go, so
	// we can count how many are open at once.
	const workers = 3
	var running, peak, arrived atomic.Int32
	release := make(chan struct{})
	three := make(chan struct{})
	h := func(w http.ResponseWriter, r *http.Request) {
		n := running.Add(1)
		defer running.Add(-1)
		for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
		}
		if arrived.Add(1) == workers {
			close(three)
		}
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, `[{"item":"unga 2kg","price":230}]`)
	}
	var shops []Shop
	for i := range 12 {
		shops = append(shops, Shop{fmt.Sprintf("shop-%02d", i), serve(t, h)})
	}

	done := make(chan error, 1)
	go func() {
		cfg := quick()
		cfg.Workers = workers
		_, err := Watch(context.Background(), http.DefaultClient, shops, cfg)
		done <- err
	}()

	select {
	case <-three:
	case err := <-done:
		t.Fatalf("Watch returned %v before %d requests were ever open at once", err, workers)
	case <-time.After(2 * time.Second):
		t.Fatalf("only %d requests were open at once after 2s, want %d. Start %d workers", running.Load(), workers, workers)
	}
	// Give extra goroutines a moment to show themselves. Correct code passes either way;
	// this only catches code that starts more than the limit.
	time.Sleep(50 * time.Millisecond)
	close(release)

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch never returned")
	}
	if p := peak.Load(); p > workers {
		t.Errorf("%d requests were open at once with Workers: %d, want at most %d", p, workers, workers)
	}
}

func TestWatchTimeout(t *testing.T) {
	// One shop never answers. Its request must give up after cfg.Timeout and the shop
	// is listed as failed; the others are still read.
	stuck := serve(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done(): // the client gave up
		case <-time.After(3 * time.Second): // a failure guard, so a missing timeout can't hang the test
		}
	})
	shops := []Shop{
		{"Slow Shop", stuck},
		{"Fast Shop", serve(t, prices(`[{"item":"unga 2kg","price":230}]`))},
	}
	cfg := quick()
	cfg.Timeout = 100 * time.Millisecond
	cfg.Retries = 3 // a timeout is not a 5xx: no retries

	start := time.Now()
	got, err := Watch(context.Background(), http.DefaultClient, shops, cfg)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Watch error = %v, want nil", err)
	}
	if !slices.Equal(got.Failed, []string{"Slow Shop"}) {
		t.Errorf("Failed = %q, want [\"Slow Shop\"]", got.Failed)
	}
	if _, ok := got.Cheapest["unga 2kg"]; !ok {
		t.Errorf("Cheapest = %v, want the fast shop's unga", got.Cheapest)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Watch took %v with a 100ms timeout. Give each request a context.WithTimeout", elapsed.Round(time.Millisecond))
	}
}

// flaky answers with the given statuses in turn, then 200 with body.
func flaky(t *testing.T, calls *atomic.Int32, statuses []int, body string) string {
	return serve(t, func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		if n <= len(statuses) {
			http.Error(w, "try later", statuses[n-1])
			return
		}
		fmt.Fprint(w, body)
	})
}

func TestWatchRetries(t *testing.T) {
	tests := []struct {
		name      string
		statuses  []int
		retries   int
		wantCalls int
		wantWaits []time.Duration
		wantFail  bool
	}{
		{"recovers after two 503s", []int{503, 503}, 3, 3, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, false},
		{"gives up after every retry", []int{500, 502, 503, 500, 500}, 2, 3, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, true},
		{"no retries allowed", []int{500}, 0, 1, nil, true},
		{"a 4xx is not retried", []int{404}, 3, 1, nil, true},
		{"a 429 is not retried either", []int{429}, 3, 1, nil, true},
		{"fine first time", nil, 3, 1, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			shops := []Shop{{"Kibanda", flaky(t, &calls, tt.statuses, `[{"item":"unga 2kg","price":230}]`)}}
			s := &noSleep{}
			cfg := Config{Workers: 1, Timeout: 2 * time.Second, Retries: tt.retries, Backoff: 10 * time.Millisecond, Sleep: s.sleep}

			got, err := Watch(context.Background(), http.DefaultClient, shops, cfg)
			if err != nil {
				t.Fatalf("Watch error = %v, want nil", err)
			}
			if n := int(calls.Load()); n != tt.wantCalls {
				t.Errorf("the shop got %d requests, want %d", n, tt.wantCalls)
			}
			if !slices.Equal(s.waits, tt.wantWaits) {
				t.Errorf("backoff waits = %v, want %v (Backoff, then doubling)", s.waits, tt.wantWaits)
			}
			failed := slices.Equal(got.Failed, []string{"Kibanda"})
			if failed != tt.wantFail {
				t.Errorf("Failed = %q, want the shop failed: %v", got.Failed, tt.wantFail)
			}
		})
	}
}

// countingLimiter is a fake rate limiter that counts its Wait calls.
type countingLimiter struct {
	calls atomic.Int32
	err   error
}

func (l *countingLimiter) Wait(ctx context.Context) error {
	l.calls.Add(1)
	return l.err
}

func TestWatchRateLimit(t *testing.T) {
	var calls atomic.Int32
	shops := []Shop{
		{"A", serve(t, prices(`[{"item":"unga 2kg","price":230}]`))},
		{"B", serve(t, prices(`[{"item":"unga 2kg","price":240}]`))},
		{"C", flaky(t, &calls, []int{503}, `[{"item":"unga 2kg","price":250}]`)},
	}
	lim := &countingLimiter{}
	cfg := quick()
	cfg.Retries = 1
	cfg.Limiter = lim
	if _, err := Watch(context.Background(), http.DefaultClient, shops, cfg); err != nil {
		t.Fatalf("Watch error = %v, want nil", err)
	}
	if n := lim.calls.Load(); n != 4 {
		t.Errorf("Limiter.Wait was called %d times, want 4: once before every request, retries included", n)
	}
}

func TestWatchLimiterRefuses(t *testing.T) {
	var hits atomic.Int32
	url := serve(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		fmt.Fprint(w, `[]`)
	})
	cfg := quick()
	cfg.Limiter = &countingLimiter{err: errors.New("rate: Wait(n=1) would exceed context deadline")}
	got, err := Watch(context.Background(), http.DefaultClient, []Shop{{"A", url}}, cfg)
	if err != nil {
		t.Fatalf("Watch error = %v, want nil", err)
	}
	if hits.Load() != 0 {
		t.Error("the shop was called although the limiter said no")
	}
	if !slices.Equal(got.Failed, []string{"A"}) {
		t.Errorf("Failed = %q, want [\"A\"]", got.Failed)
	}
}

func TestWatchCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the user quit before we started
	shops := []Shop{{"A", serve(t, prices(`[{"item":"unga 2kg","price":230}]`))}}
	_, err := Watch(ctx, http.DefaultClient, shops, quick())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Watch with a cancelled context = %v, want context.Canceled", err)
	}
}

func TestReport(t *testing.T) {
	tests := []struct {
		name string
		s    Summary
		want string
	}{
		{"prices and a failure", Summary{
			Cheapest: map[string]Offer{
				"unga 2kg":   {"Mama Njeri Shop", 230},
				"sukari 1kg": {"Duka la Kona", 180},
			},
			Failed: []string{"Gikomba Wholesale", "Ngara Stores"},
		}, "Cheapest prices:\nsukari 1kg: Ksh180 at Duka la Kona\nunga 2kg: Ksh230 at Mama Njeri Shop\nFailed: Gikomba Wholesale, Ngara Stores\n"},
		{"no failures", Summary{Cheapest: map[string]Offer{"chumvi 500g": {"Duka la Kona", 30}}},
			"Cheapest prices:\nchumvi 500g: Ksh30 at Duka la Kona\n"},
		{"nothing at all", Summary{}, "Cheapest prices:\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Report(tt.s); got != tt.want {
				t.Errorf("Report\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
