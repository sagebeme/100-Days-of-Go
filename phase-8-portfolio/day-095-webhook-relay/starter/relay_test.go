package webhookrelay

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// sent is one request the fake client received.
type sent struct {
	url    string
	method string
	header http.Header
	body   string
}

// fakeClient answers with the scripted statuses in order (0 means a network error),
// then 200 for ever. It records every request.
type fakeClient struct {
	mu       sync.Mutex
	statuses []int
	reqs     []sent
}

func (f *fakeClient) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqs = append(f.reqs, sent{r.URL.String(), r.Method, r.Header.Clone(), string(b)})
	status := http.StatusOK
	if len(f.statuses) > 0 {
		status, f.statuses = f.statuses[0], f.statuses[1:]
	}
	if status == 0 {
		return nil, errors.New("dial tcp: connection refused")
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("ok"))}, nil
}

// fakeSleep records the waits instead of waiting.
type fakeSleep struct {
	mu    sync.Mutex
	waits []time.Duration
}

func (f *fakeSleep) Sleep(ctx context.Context, d time.Duration) error {
	f.mu.Lock()
	f.waits = append(f.waits, d)
	f.mu.Unlock()
	return ctx.Err()
}

var subA = Subscriber{URL: "https://ledger.example/hooks", Secret: []byte("sub-a-secret")}
var subB = Subscriber{URL: "https://sms.example/hooks", Secret: []byte("sub-b-secret")}

func newRelay(client *fakeClient, sl *fakeSleep, change func(*Config)) *Relay {
	cfg := Config{
		Secret:      secret,
		Subscribers: []Subscriber{subA},
		Client:      client,
		Now:         func() time.Time { return t0 },
		Sleep:       sl.Sleep,
		BaseDelay:   time.Second,
		MaxDelay:    time.Minute,
	}
	if change != nil {
		change(&cfg)
	}
	return New(cfg)
}

func signed(b string, at time.Time) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(SignatureHeader, Sign(secret, at, []byte(b)))
	return r
}

func call(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHandler(t *testing.T) {
	rl := newRelay(&fakeClient{}, &fakeSleep{}, nil)
	steps := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   string
	}{
		{"new event", signed(string(body), t0), http.StatusAccepted, `"queued"`},
		{"same event again", signed(string(body), t0), http.StatusOK, `"duplicate"`},
		{"another event", signed(`{"id":"evt_002"}`, t0), http.StatusAccepted, `"queued"`},
		{"bad signature", func() *http.Request {
			r := signed(`{"id":"evt_003"}`, t0)
			r.Header.Set(SignatureHeader, Sign([]byte("guess"), t0, []byte(`{"id":"evt_003"}`)))
			return r
		}(), http.StatusUnauthorized, `"error"`},
		{"no signature", httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(`{"id":"evt_004"}`)), http.StatusUnauthorized, `"error"`},
		{"signed ten minutes ago", signed(`{"id":"evt_005"}`, t0.Add(-10*time.Minute)), http.StatusUnauthorized, `"error"`},
		{"signed but not JSON", signed(`id=evt_006`, t0), http.StatusBadRequest, `"error"`},
		{"signed JSON with no id", signed(`{"type":"payment.received"}`, t0), http.StatusBadRequest, `"error"`},
		{"GET", httptest.NewRequest(http.MethodGet, "/webhooks", nil), http.StatusMethodNotAllowed, `"error"`},
		{"huge body", signed(`{"id":"x","pad":"`+strings.Repeat("a", MaxBodyBytes)+`"}`, t0), http.StatusRequestEntityTooLarge, `"error"`},
	}
	for _, s := range steps {
		w := call(rl, s.req)
		if w.Code != s.wantStatus || !strings.Contains(w.Body.String(), s.wantBody) {
			t.Errorf("%s: status %d, body %s; want %d and %s", s.name, w.Code, strings.TrimSpace(w.Body.String()), s.wantStatus, s.wantBody)
		}
	}
	// A rejected request must not use up its ID: evt_003 with a good signature gets in.
	if w := call(rl, signed(`{"id":"evt_003"}`, t0)); w.Code != http.StatusAccepted {
		t.Errorf("evt_003 after its forged copy was refused: status %d, want 202", w.Code)
	}
}

func TestHandlerQueueFull(t *testing.T) {
	rl := newRelay(&fakeClient{}, &fakeSleep{}, func(c *Config) { c.QueueSize = 1 })
	if w := call(rl, signed(`{"id":"evt_a"}`, t0)); w.Code != http.StatusAccepted {
		t.Fatalf("first event: status %d, want 202", w.Code)
	}
	for i := range 2 {
		w := call(rl, signed(`{"id":"evt_b"}`, t0))
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("evt_b try %d with a full queue: status %d, want 503 (not remembered, so not a duplicate)", i+1, w.Code)
		}
		if w.Header().Get("Retry-After") == "" {
			t.Errorf("503 should carry a Retry-After header")
		}
	}
}

func TestDeliverRetries(t *testing.T) {
	tests := []struct {
		name         string
		statuses     []int
		maxAttempts  int
		wantAttempts int
		wantStatus   int
		wantOK       bool
		wantWaits    []time.Duration
	}{
		{"first try", nil, 5, 1, 200, true, nil},
		{"two failures then success", []int{500, 502}, 5, 3, 200, true, []time.Duration{time.Second, 2 * time.Second}},
		{"network error then success", []int{0}, 5, 2, 200, true, []time.Duration{time.Second}},
		{"429 is retried", []int{429}, 5, 2, 200, true, []time.Duration{time.Second}},
		{"204 counts as delivered", []int{204}, 5, 1, 204, true, nil},
		{"404 is not retried", []int{404}, 5, 1, 404, false, nil},
		{"410 is not retried", []int{410}, 5, 1, 410, false, nil},
		{"gives up after MaxAttempts", []int{503, 503, 503, 503, 503}, 4, 4, 503, false, []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeClient{statuses: slices.Clone(tt.statuses)}
			sl := &fakeSleep{}
			rl := newRelay(client, sl, func(c *Config) { c.MaxAttempts = tt.maxAttempts })
			res := rl.Deliver(context.Background(), subA, Event{ID: "evt_001"}, body)
			if res.Attempts != tt.wantAttempts || res.Status != tt.wantStatus || (res.Err == nil) != tt.wantOK {
				t.Errorf("Result = %d attempts, status %d, err %v; want %d attempts, status %d, delivered %v",
					res.Attempts, res.Status, res.Err, tt.wantAttempts, tt.wantStatus, tt.wantOK)
			}
			if res.EventID != "evt_001" || res.URL != subA.URL {
				t.Errorf("Result names event %q at %q, want evt_001 at %s", res.EventID, res.URL, subA.URL)
			}
			if !slices.Equal(sl.waits, tt.wantWaits) {
				t.Errorf("waited %v, want %v", sl.waits, tt.wantWaits)
			}
			if len(client.reqs) != tt.wantAttempts {
				t.Errorf("client got %d requests, want %d", len(client.reqs), tt.wantAttempts)
			}
		})
	}
}

func TestDeliverStopsWhenCancelled(t *testing.T) {
	client := &fakeClient{statuses: []int{500, 500, 500}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rl := newRelay(client, &fakeSleep{}, nil)
	res := rl.Deliver(ctx, subA, Event{ID: "evt_001"}, body)
	if !errors.Is(res.Err, context.Canceled) {
		t.Errorf("Err = %v, want context.Canceled", res.Err)
	}
	if res.Attempts > 1 {
		t.Errorf("made %d attempts after the context was cancelled, want at most 1", res.Attempts)
	}
}

func TestDeliverSignsForTheSubscriber(t *testing.T) {
	client := &fakeClient{}
	rl := newRelay(client, &fakeSleep{}, nil)
	rl.Deliver(context.Background(), subA, Event{ID: "evt_001"}, body)
	if len(client.reqs) != 1 {
		t.Fatalf("client got %d requests, want 1", len(client.reqs))
	}
	r := client.reqs[0]
	if r.method != http.MethodPost || r.url != subA.URL || r.body != string(body) {
		t.Errorf("sent %s %s with body %q; want POST %s with the original body", r.method, r.url, r.body, subA.URL)
	}
	if ct := r.header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if id := r.header.Get("Webhook-Id"); id != "evt_001" {
		t.Errorf("Webhook-Id = %q, want evt_001, so the subscriber can dedupe too", id)
	}
	if err := Verify(subA.Secret, r.header.Get(SignatureHeader), []byte(r.body), t0, time.Minute); err != nil {
		t.Errorf("the subscriber can't verify the forwarded request with its own secret: %v", err)
	}
}

func TestRunForwardsToEverySubscriber(t *testing.T) {
	client := &fakeClient{}
	results := make(chan Result, 4)
	rl := newRelay(client, &fakeSleep{}, func(c *Config) {
		c.Subscribers = []Subscriber{subA, subB}
		c.OnResult = func(r Result) { results <- r }
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		rl.Run(ctx)
		close(done)
	}()

	if w := call(rl, signed(string(body), t0)); w.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202", w.Code)
	}
	var urls []string
	for range 2 {
		select {
		case r := <-results:
			if r.Err != nil {
				t.Errorf("delivery to %s failed: %v", r.URL, r.Err)
			}
			urls = append(urls, r.URL)
		case <-time.After(2 * time.Second):
			t.Fatal("Run delivered nothing within 2 seconds")
		}
	}
	slices.Sort(urls)
	if want := []string{subA.URL, subB.URL}; !slices.Equal(urls, want) {
		t.Errorf("delivered to %v, want %v", urls, want)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run didn't return after the context was cancelled")
	}
}
