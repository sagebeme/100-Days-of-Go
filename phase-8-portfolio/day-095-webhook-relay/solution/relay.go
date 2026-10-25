package webhookrelay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Doer sends an HTTP request. *http.Client is one; tests pass a fake.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Subscriber is one place events are forwarded to. Each has its own secret, so the
// relay signs what it forwards and the subscriber can check it came from the relay.
type Subscriber struct {
	URL    string
	Secret []byte
}

// Event is the part of a webhook body the relay reads. The rest is forwarded untouched.
type Event struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Result is the outcome of delivering one event to one subscriber.
type Result struct {
	EventID  string
	URL      string
	Attempts int
	Status   int   // the last HTTP status, 0 if there was none
	Err      error // nil when delivered
}

// Config holds everything the relay needs. Zero values get the defaults shown.
type Config struct {
	Secret      []byte // the inbound secret, shared with the webhook sender (required)
	Subscribers []Subscriber
	Client      Doer                                             // default: an http.Client with a 10 s timeout
	Now         func() time.Time                                 // default: time.Now
	Sleep       func(ctx context.Context, d time.Duration) error // default: waits, or returns ctx.Err()
	OnResult    func(Result)                                     // called after each delivery; default: nothing
	Tolerance   time.Duration                                    // default: 5 minutes
	MaxAttempts int                                              // default: 5
	BaseDelay   time.Duration                                    // default: 1 second
	MaxDelay    time.Duration                                    // default: 1 minute
	DedupeTTL   time.Duration                                    // default: 24 hours
	QueueSize   int                                              // default: 100
}

// MaxBodyBytes caps an incoming webhook.
const MaxBodyBytes = 1 << 20

// Relay is an http.Handler that accepts webhooks, plus a worker (Run) that forwards them.
type Relay struct {
	cfg   Config
	seen  *Deduper
	queue chan job
}

type job struct {
	event Event
	body  []byte
}

// New returns a relay with the defaults filled in.
func New(cfg Config) *Relay {
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleep
	}
	if cfg.OnResult == nil {
		cfg.OnResult = func(Result) {}
	}
	if cfg.Tolerance <= 0 {
		cfg.Tolerance = 5 * time.Minute
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = time.Second
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = time.Minute
	}
	if cfg.DedupeTTL <= 0 {
		cfg.DedupeTTL = 24 * time.Hour
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	return &Relay{cfg: cfg, seen: NewDeduper(cfg.DedupeTTL), queue: make(chan job, cfg.QueueSize)}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ServeHTTP accepts one webhook:
//
//	405  not a POST
//	413  body over MaxBodyBytes
//	401  bad, missing or stale signature
//	400  body isn't JSON with a non-empty "id"
//	200  {"status":"duplicate"}  an ID already seen within DedupeTTL
//	503  the queue is full (the ID is not remembered, so the sender's retry gets in)
//	202  {"status":"queued"}
func (rl *Relay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "can't read body"})
		return
	}
	if err := Verify(rl.cfg.Secret, r.Header.Get(SignatureHeader), body, rl.cfg.Now(), rl.cfg.Tolerance); err != nil {
		// Don't say which check failed: that helps an attacker more than a sender.
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}
	var ev Event
	if err := json.Unmarshal(body, &ev); err != nil || ev.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `body must be JSON with an "id"`})
		return
	}
	if rl.seen.Seen(ev.ID, rl.cfg.Now()) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
		return
	}
	select {
	case rl.queue <- job{ev, body}:
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
	default:
		rl.seen.Forget(ev.ID)
		w.Header().Set("Retry-After", "30")
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "busy, try again"})
	}
}

// Run forwards queued events to every subscriber until ctx is cancelled.
func (rl *Relay) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-rl.queue:
			for _, sub := range rl.cfg.Subscribers {
				rl.cfg.OnResult(rl.Deliver(ctx, sub, j.event, j.body))
			}
		}
	}
}

// Deliver POSTs body to sub, signed with sub.Secret, and retries until it gets a 2xx,
// MaxAttempts is reached, the error can't be fixed by retrying (any other 4xx), or ctx
// ends. Before attempt n+1 it sleeps Backoff(n, BaseDelay, MaxDelay).
func (rl *Relay) Deliver(ctx context.Context, sub Subscriber, ev Event, body []byte) Result {
	res := Result{EventID: ev.ID, URL: sub.URL}
	for attempt := 1; attempt <= rl.cfg.MaxAttempts; attempt++ {
		res.Attempts = attempt
		status, err := rl.send(ctx, sub, ev, body)
		res.Status = status
		if err == nil && status >= 200 && status < 300 {
			res.Err = nil
			return res
		}
		if err != nil {
			res.Err = err
		} else {
			res.Err = fmt.Errorf("webhookrelay: %s answered %d", sub.URL, status)
		}
		if err == nil && !Retryable(status) {
			return res
		}
		if attempt == rl.cfg.MaxAttempts {
			break
		}
		if err := rl.cfg.Sleep(ctx, Backoff(attempt, rl.cfg.BaseDelay, rl.cfg.MaxDelay)); err != nil {
			res.Err = err
			return res
		}
	}
	return res
}

func (rl *Relay) send(ctx context.Context, sub Subscriber, ev Event, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Webhook-Id", ev.ID)
	req.Header.Set(SignatureHeader, Sign(sub.Secret, rl.cfg.Now(), body))
	resp, err := rl.cfg.Client.Do(req)
	if err != nil {
		return 0, err
	}
	// Drain a little and close, so the connection can be reused.
	io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	resp.Body.Close()
	return resp.StatusCode, nil
}

// Retryable reports whether a status is worth another try: 408, 429 and every 5xx.
func Retryable(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500
}

// Backoff returns the wait after attempt number attempt (from 1): base, 2×base,
// 4×base and so on, never more than max.
func Backoff(attempt int, base, max time.Duration) time.Duration {
	d := base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= max {
			return max
		}
	}
	return min(d, max)
}

// Deduper remembers event IDs for a while. It is safe for concurrent use.
type Deduper struct {
	mu    sync.Mutex
	ttl   time.Duration
	at    map[string]time.Time
	order []seen // in the order they were added, so oldest first
}

type seen struct {
	id string
	at time.Time
}

// NewDeduper remembers each ID for ttl.
func NewDeduper(ttl time.Duration) *Deduper {
	return &Deduper{ttl: ttl, at: map[string]time.Time{}}
}

// Seen reports whether id was seen in the last ttl. If not, it remembers id from now.
// Times passed to Seen must not go backwards.
func (d *Deduper) Seen(id string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prune(now)
	if _, ok := d.at[id]; ok {
		return true
	}
	d.at[id] = now
	d.order = append(d.order, seen{id, now})
	return false
}

// Forget drops id, so the next Seen(id) is false.
func (d *Deduper) Forget(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.at, id)
}

// Len returns how many IDs are remembered.
func (d *Deduper) Len(now time.Time) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prune(now)
	return len(d.at)
}

// prune drops IDs older than ttl. They were added in time order, so the old ones are
// at the front: pruning costs O(1) per ID over its whole life.
func (d *Deduper) prune(now time.Time) {
	i := 0
	for ; i < len(d.order) && now.Sub(d.order[i].at) >= d.ttl; i++ {
		e := d.order[i]
		if d.at[e.id].Equal(e.at) { // not forgotten and added again since
			delete(d.at, e.id)
		}
	}
	d.order = d.order[i:]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
