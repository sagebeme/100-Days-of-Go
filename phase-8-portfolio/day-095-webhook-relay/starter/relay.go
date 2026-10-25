package webhookrelay

import (
	"context"
	"net/http"
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
// Add the fields you need: the config, a Deduper, a buffered channel of queued events.
type Relay struct {
}

// New returns a relay with the defaults filled in.
func New(cfg Config) *Relay {
	// Your code here
	return &Relay{}
}

// ServeHTTP accepts one webhook. Every response body is JSON:
//
//	405  not a POST
//	413  body over MaxBodyBytes
//	401  bad, missing or stale signature (don't say which)
//	400  body isn't JSON with a non-empty "id"
//	200  {"status":"duplicate"}  an ID already seen within DedupeTTL
//	503  the queue is full, with Retry-After; the ID is not remembered
//	202  {"status":"queued"}
func (rl *Relay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Your code here
}

// Run forwards queued events to every subscriber, calling OnResult after each
// delivery, until ctx is cancelled.
func (rl *Relay) Run(ctx context.Context) {
	// Your code here
}

// Deliver POSTs body to sub with Content-Type application/json, a Webhook-Id header
// and a signature made with sub.Secret. It retries until it gets a 2xx, MaxAttempts is
// reached, the status isn't Retryable, or ctx ends. Before attempt n+1 it calls
// Sleep(ctx, Backoff(n, BaseDelay, MaxDelay)).
func (rl *Relay) Deliver(ctx context.Context, sub Subscriber, ev Event, body []byte) Result {
	// Your code here
	return Result{}
}

// Retryable reports whether a status is worth another try: 408, 429 and every 5xx.
func Retryable(status int) bool {
	// Your code here
	return false
}

// Backoff returns the wait after attempt number attempt (from 1): base, 2×base,
// 4×base and so on, never more than max. It must not overflow for large attempts.
func Backoff(attempt int, base, max time.Duration) time.Duration {
	// Your code here
	return 0
}

// Deduper remembers event IDs for a while. It is safe for concurrent use.
type Deduper struct {
}

// NewDeduper remembers each ID for ttl.
func NewDeduper(ttl time.Duration) *Deduper {
	// Your code here
	return &Deduper{}
}

// Seen reports whether id was seen in the last ttl. If not, it remembers id from now.
// Times passed to Seen never go backwards. Drop expired IDs so memory doesn't grow.
func (d *Deduper) Seen(id string, now time.Time) bool {
	// Your code here
	return false
}

// Forget drops id, so the next Seen(id) is false.
func (d *Deduper) Forget(id string) {
	// Your code here
}

// Len returns how many IDs are remembered and not yet expired at now.
func (d *Deduper) Len(now time.Time) int {
	// Your code here
	return 0
}
