package backgroundjobs

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

// fakeClock is a Clock the test moves by hand.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

var start = time.Date(2026, 10, 24, 7, 0, 0, 0, time.UTC)

func fixed(v float64) func() float64 { return func() float64 { return v } }

func TestBackoff(t *testing.T) {
	tests := []struct {
		attempt   int
		base, max time.Duration
		rnd       float64
		want      time.Duration
	}{
		{1, time.Second, time.Minute, 0, 500 * time.Millisecond},
		{1, time.Second, time.Minute, 0.5, 750 * time.Millisecond},
		{2, time.Second, time.Minute, 0, time.Second},
		{3, time.Second, time.Minute, 0.5, 3 * time.Second},
		{6, time.Second, time.Minute, 0, 16 * time.Second},
		{7, time.Second, time.Minute, 0, 30 * time.Second},   // ceiling 64s, capped at 60s
		{100, time.Second, time.Minute, 0, 30 * time.Second}, // no overflow
		{100, time.Second, time.Minute, 0.999, 59970 * time.Millisecond},
	}
	for _, tt := range tests {
		if got := Backoff(tt.attempt, tt.base, tt.max, fixed(tt.rnd)); got != tt.want {
			t.Errorf("Backoff(%d, %v, %v, rnd=%v) = %v, want %v", tt.attempt, tt.base, tt.max, tt.rnd, got, tt.want)
		}
	}
}

func TestBackoffStaysInRange(t *testing.T) {
	for attempt := 1; attempt <= 10; attempt++ {
		for _, r := range []float64{0, 0.25, 0.5, 0.75, 0.9999} {
			got := Backoff(attempt, 100*time.Millisecond, 10*time.Second, fixed(r))
			ceiling := min(100*time.Millisecond<<(attempt-1), 10*time.Second)
			if got < ceiling/2 || got > ceiling {
				t.Errorf("Backoff(%d, rnd=%v) = %v, want between %v and %v", attempt, r, got, ceiling/2, ceiling)
			}
		}
	}
}

// recorder is a handler that fails each key a set number of times, then succeeds.
type recorder struct {
	mu       sync.Mutex
	failures map[string]int
	ran      []string
}

func (r *recorder) handle(ctx context.Context, j Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ran = append(r.ran, j.Key)
	if r.failures[j.Key] > 0 {
		r.failures[j.Key]--
		return errors.New("safaricom gateway timeout")
	}
	return nil
}

func newQueue(failures map[string]int) (*Queue, *fakeClock, *recorder) {
	clock := &fakeClock{now: start}
	rec := &recorder{failures: failures}
	q := NewQueue(clock, rec.handle, Options{MaxAttempts: 3, BaseDelay: time.Second, MaxDelay: time.Minute, Rand: fixed(0)})
	return q, clock, rec
}

func TestEnqueueDedupes(t *testing.T) {
	q, _, _ := newQueue(nil)
	if err := q.Enqueue("sms:1042", "Seat 3 on 46K"); err != nil {
		t.Fatalf("first Enqueue = %v, want nil", err)
	}
	if err := q.Enqueue("sms:1042", "Seat 3 on 46K"); !errors.Is(err, ErrDuplicate) {
		t.Errorf("same key again = %v, want ErrDuplicate", err)
	}
	q.RunDue(context.Background())
	if err := q.Enqueue("sms:1042", "again"); !errors.Is(err, ErrDuplicate) {
		t.Errorf("key of a finished job = %v, want ErrDuplicate", err)
	}
	if got := len(q.Pending()); got != 0 {
		t.Errorf("%d jobs pending, want 0", got)
	}
}

func TestRunDueOrderAndCount(t *testing.T) {
	q, _, rec := newQueue(nil)
	for _, k := range []string{"a", "b", "c"} {
		q.Enqueue(k, "")
	}
	if n := q.RunDue(context.Background()); n != 3 {
		t.Errorf("RunDue = %d, want 3", n)
	}
	if want := []string{"a", "b", "c"}; !slices.Equal(rec.ran, want) {
		t.Errorf("ran %v, want %v (enqueue order)", rec.ran, want)
	}
	if n := q.RunDue(context.Background()); n != 0 {
		t.Errorf("second RunDue = %d, want 0: nothing is left", n)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	q, clock, rec := newQueue(map[string]int{"sms:7": 2})
	ctx := context.Background()
	q.Enqueue("sms:7", "")

	q.RunDue(ctx) // fails, attempt 1, waits 500ms
	p := q.Pending()
	if len(p) != 1 || p[0].Attempts != 1 || !p[0].RunAt.Equal(start.Add(500*time.Millisecond)) || p[0].LastErr != "safaricom gateway timeout" {
		t.Fatalf("after one failure Pending = %+v, want 1 job, Attempts 1, RunAt +500ms, LastErr set", p)
	}
	clock.Advance(499 * time.Millisecond)
	if n := q.RunDue(ctx); n != 0 {
		t.Errorf("RunDue 1ms early ran %d jobs, want 0", n)
	}
	clock.Advance(time.Millisecond)
	q.RunDue(ctx) // fails, attempt 2, waits 1s from now
	p = q.Pending()
	if len(p) != 1 || p[0].Attempts != 2 || !p[0].RunAt.Equal(start.Add(1500*time.Millisecond)) {
		t.Fatalf("after two failures Pending = %+v, want Attempts 2, RunAt +1.5s", p)
	}
	clock.Advance(time.Second)
	q.RunDue(ctx) // succeeds
	if len(q.Pending()) != 0 || len(q.DeadLetters()) != 0 {
		t.Errorf("after success: pending %v, dead %v, want both empty", q.Pending(), q.DeadLetters())
	}
	if len(rec.ran) != 3 {
		t.Errorf("handler ran %d times, want 3", len(rec.ran))
	}
}

func TestDeadLetter(t *testing.T) {
	q, clock, _ := newQueue(map[string]int{"sms:bad": 99, "sms:ok": 0})
	ctx := context.Background()
	q.Enqueue("sms:bad", "")
	q.Enqueue("sms:ok", "")
	for range 5 {
		q.RunDue(ctx)
		clock.Advance(time.Minute)
	}
	dead := q.DeadLetters()
	if len(dead) != 1 || dead[0].Key != "sms:bad" || dead[0].Attempts != 3 {
		t.Fatalf("DeadLetters = %+v, want sms:bad after 3 attempts", dead)
	}
	if len(q.Pending()) != 0 {
		t.Errorf("Pending = %+v, want empty: a dead job must not be retried", q.Pending())
	}
}

func TestPermanentErrorSkipsRetries(t *testing.T) {
	clock := &fakeClock{now: start}
	runs := 0
	h := func(ctx context.Context, j Job) error {
		runs++
		return Permanent(errors.New("no such phone number"))
	}
	q := NewQueue(clock, h, Options{MaxAttempts: 5, BaseDelay: time.Second, MaxDelay: time.Minute, Rand: fixed(0)})
	q.Enqueue("sms:x", "")
	q.RunDue(context.Background())
	if dead := q.DeadLetters(); len(dead) != 1 || dead[0].Attempts != 1 || dead[0].LastErr != "no such phone number" {
		t.Errorf("DeadLetters = %+v, want sms:x after 1 attempt", dead)
	}
	if runs != 1 {
		t.Errorf("handler ran %d times, want 1", runs)
	}
}

func TestRetryWaitsBehindDueJobs(t *testing.T) {
	q, clock, rec := newQueue(map[string]int{"a": 1})
	ctx := context.Background()
	q.Enqueue("a", "")
	q.RunDue(ctx) // a fails, due at +500ms
	clock.Advance(200 * time.Millisecond)
	q.Enqueue("b", "") // due at +200ms
	clock.Advance(time.Second)
	q.RunDue(ctx)
	if want := []string{"a", "b", "a"}; !slices.Equal(rec.ran, want) {
		t.Errorf("ran %v, want %v: the earliest RunAt goes first", rec.ran, want)
	}
}

func TestRun(t *testing.T) {
	clock := &fakeClock{now: start}
	ran := make(chan string, 1)
	q := NewQueue(clock, func(ctx context.Context, j Job) error { ran <- j.Key; return nil }, Options{MaxAttempts: 3, BaseDelay: time.Second, MaxDelay: time.Minute, Rand: fixed(0)})
	q.Enqueue("sms:1", "")
	tick := make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() { q.Run(ctx, tick); close(stopped) }()
	tick <- start
	select {
	case k := <-ran:
		if k != "sms:1" {
			t.Errorf("ran %q, want sms:1", k)
		}
	case <-time.After(time.Second):
		t.Fatal("a tick didn't run the due job")
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Run didn't return after ctx was cancelled")
	}
}

func TestIdempotent(t *testing.T) {
	sent := map[string]int{}
	fail := true
	h := Idempotent(func(ctx context.Context, j Job) error {
		if fail {
			fail = false
			return errors.New("timeout")
		}
		sent[j.Key]++
		return nil
	})
	ctx := context.Background()
	j := Job{Key: "sms:1042"}
	if err := h(ctx, j); err == nil {
		t.Fatal("first call should pass the error through")
	}
	for range 3 {
		if err := h(ctx, j); err != nil {
			t.Errorf("h = %v, want nil", err)
		}
	}
	h(ctx, Job{Key: "sms:2000"})
	if sent["sms:1042"] != 1 || sent["sms:2000"] != 1 {
		t.Errorf("sent %v, want each key once: a failed run doesn't count, a repeat after success does nothing", sent)
	}
}

func TestQueueIsSafeForGoroutines(t *testing.T) {
	q, _, _ := newQueue(nil)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() { q.Enqueue(string(rune('A'+i%26))+"-key", "") })
		wg.Go(func() { q.RunDue(context.Background()) })
	}
	wg.Wait()
	q.RunDue(context.Background())
	if len(q.Pending()) != 0 {
		t.Errorf("Pending = %d jobs, want 0", len(q.Pending()))
	}
}
