package ratelimiting

import (
	"context"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/time/rate"
)

// t0 is any fixed moment. The bucket tests pass times explicitly, so there's no clock to
// wait for: the test decides what "now" is.
var t0 = time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

func ms(n int) time.Time { return t0.Add(time.Duration(n) * time.Millisecond) }

func TestThrottle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tick := make(chan time.Time)
		var mu sync.Mutex
		var sent []string
		send := func(n string) {
			mu.Lock()
			sent = append(sent, n)
			mu.Unlock()
		}
		count := func() int {
			mu.Lock()
			defer mu.Unlock()
			return len(sent)
		}

		numbers := []string{"0712000001", "0712000002", "0712000003"}
		done := make(chan error, 1)
		go func() { done <- Throttle(context.Background(), numbers, tick, send) }()

		// synctest.Wait returns once Throttle is stuck waiting, so we can look inside.
		synctest.Wait()
		if n := count(); n != 0 {
			t.Fatalf("%d messages went out before the first tick, want 0. Wait for a tick before each send", n)
		}
		for i := range numbers {
			select {
			case tick <- time.Now():
			case err := <-done:
				t.Fatalf("Throttle returned %v after %d ticks, want it to keep waiting for ticks", err, i)
			}
			synctest.Wait()
			if n := count(); n != i+1 {
				t.Fatalf("after %d ticks, %d messages were sent, want %d", i+1, n, i+1)
			}
		}
		if err := <-done; err != nil {
			t.Errorf("Throttle = %v, want nil", err)
		}
		if !slices.Equal(sent, numbers) {
			t.Errorf("sent %q, want %q in order", sent, numbers)
		}
	})
}

func TestThrottleCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		tick := make(chan time.Time)
		var sent atomic.Int32
		done := make(chan error, 1)
		go func() {
			done <- Throttle(ctx, []string{"a", "b", "c", "d"}, tick, func(string) { sent.Add(1) })
		}()

		select {
		case tick <- time.Now():
		case err := <-done:
			t.Fatalf("Throttle returned %v before the first tick, want it to wait for ticks", err)
		}
		synctest.Wait()
		cancel() // the user stops the campaign after one message
		synctest.Wait()

		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("Throttle after cancel = %v, want context.Canceled", err)
			}
		default:
			t.Fatal("Throttle is still waiting for a tick after its context was cancelled. Add a ctx.Done() case")
		}
		if n := sent.Load(); n != 1 {
			t.Errorf("%d messages were sent, want 1", n)
		}
	})
}

func TestThrottleNothingToSend(t *testing.T) {
	// No items: return at once, without waiting for a tick.
	if err := Throttle(context.Background(), nil, make(chan time.Time), func(string) {}); err != nil {
		t.Errorf("Throttle(no items) = %v, want nil", err)
	}
}

func TestTokenBucket(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		every    time.Duration
		at       []int  // milliseconds after t0
		want     []bool // what Allow should say at each time
	}{
		{"burst, then wait", 2, time.Second,
			[]int{0, 0, 0, 500, 1000, 1000, 1999, 2000},
			[]bool{true, true, false, false, true, false, false, true}},
		{"refills up to capacity only", 3, time.Second,
			[]int{0, 0, 0, 60000, 60000, 60000, 60000},
			[]bool{true, true, true, true, true, true, false}},
		{"part of a step carries over", 2, time.Second,
			[]int{0, 0, 1500, 1999, 2000},
			[]bool{true, true, true, false, true}},
		{"a full bucket can't save up", 1, time.Second,
			[]int{0, 700, 1400, 1600, 2100, 2399, 2400, 2500},
			[]bool{true, false, true, false, false, false, true, false}},
		{"a full bucket restarts its clock", 2, time.Second,
			[]int{0, 5500, 5500, 6000, 6500},
			[]bool{true, true, true, false, true}},
		{"steady, one per tick", 1, 100 * time.Millisecond,
			[]int{0, 100, 200, 250, 300},
			[]bool{true, true, true, false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTokenBucket(tt.capacity, tt.every, t0)
			for i, m := range tt.at {
				if got := b.Allow(ms(m)); got != tt.want[i] {
					t.Errorf("call %d: Allow(t0+%dms) = %v, want %v", i+1, m, got, tt.want[i])
				}
			}
		})
	}
}

func TestTokenBucketConcurrent(t *testing.T) {
	// 100 goroutines hit a bucket of 10 at the same instant: exactly 10 get through.
	b := NewTokenBucket(10, time.Second, t0)
	var ok atomic.Int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if b.Allow(t0) {
				ok.Add(1)
			}
		})
	}
	wg.Wait()
	if n := ok.Load(); n != 10 {
		t.Errorf("%d of 100 calls were allowed by a bucket of 10, want exactly 10", n)
	}
}

func TestClientLimiter(t *testing.T) {
	l := NewClientLimiter(rate.Every(time.Second), 2)
	steps := []struct {
		client string
		at     int
		want   bool
	}{
		{"shop-a", 0, true},
		{"shop-a", 0, true},
		{"shop-a", 0, false},   // shop-a used its burst
		{"shop-b", 0, true},    // shop-b has its own
		{"shop-b", 0, true},    //
		{"shop-a", 999, false}, // not quite a second
		{"shop-a", 1000, true}, // one token back
		{"shop-a", 1000, false},
		{"shop-c", 1000, true},
		{"shop-b", 5000, true}, // refilled to its burst of 2, no more
		{"shop-b", 5000, true},
		{"shop-b", 5000, false},
	}
	for i, s := range steps {
		if got := l.Allow(s.client, ms(s.at)); got != s.want {
			t.Errorf("step %d: Allow(%q, t0+%dms) = %v, want %v", i+1, s.client, s.at, got, s.want)
		}
	}
}

func TestClientLimiterConcurrent(t *testing.T) {
	// 20 clients, 10 calls each, all at once. Each has a burst of 3. Run with -race.
	l := NewClientLimiter(rate.Every(time.Minute), 3)
	var ok atomic.Int32
	var wg sync.WaitGroup
	for c := range 20 {
		client := string(rune('a' + c))
		for range 10 {
			wg.Go(func() {
				if l.Allow(client, t0) {
					ok.Add(1)
				}
			})
		}
	}
	wg.Wait()
	if n := ok.Load(); n != 60 {
		t.Errorf("%d calls were allowed, want 60 (3 for each of 20 clients)", n)
	}
}
