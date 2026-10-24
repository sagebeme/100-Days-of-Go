package contextcancel

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

// Every test here runs inside synctest.Test, where time is fake. It jumps forward
// whenever every goroutine is waiting, so "after 5s" takes no real time and is exact.

// slow is a fake network call: it answers after d, or gives up when ctx is done.
func slow(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestSleep(t *testing.T) {
	tests := []struct {
		name        string
		d           time.Duration
		cancelAfter time.Duration // 0: never cancel
		deadline    time.Duration // 0: no deadline
		precancel   bool
		wantErr     error
		wantElapsed time.Duration
	}{
		{"sleeps the full time", 5 * time.Second, 0, 0, false, nil, 5 * time.Second},
		{"cancelled part way", 5 * time.Second, 2 * time.Second, 0, false, context.Canceled, 2 * time.Second},
		{"deadline comes first", 5 * time.Second, 0, time.Second, false, context.DeadlineExceeded, time.Second},
		{"deadline comes later", 3 * time.Second, 0, 10 * time.Second, false, nil, 3 * time.Second},
		{"already cancelled", 5 * time.Second, 0, 0, true, context.Canceled, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if tt.deadline > 0 {
					var cancelDeadline context.CancelFunc
					ctx, cancelDeadline = context.WithTimeout(ctx, tt.deadline)
					defer cancelDeadline()
				}
				if tt.cancelAfter > 0 {
					stop := time.AfterFunc(tt.cancelAfter, cancel)
					defer stop.Stop()
				}
				if tt.precancel {
					cancel()
				}

				start := time.Now()
				err := Sleep(ctx, tt.d)
				elapsed := time.Since(start)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Sleep(ctx, %v) = %v, want %v", tt.d, err, tt.wantErr)
				}
				if elapsed != tt.wantElapsed {
					t.Errorf("Sleep(ctx, %v) returned after %v, want %v", tt.d, elapsed, tt.wantElapsed)
				}
			})
		})
	}
}

func TestFetchFare(t *testing.T) {
	closed := errors.New("SACCO office closed")
	tests := []struct {
		name        string
		takes       time.Duration
		fails       error
		timeout     time.Duration
		want        int
		wantErr     error
		wantMsg     string
		wantElapsed time.Duration
	}{
		{"answers in time", time.Second, nil, 5 * time.Second, 100, nil, "", time.Second},
		{"too slow", 10 * time.Second, nil, 5 * time.Second, 0, context.DeadlineExceeded,
			"fare for Kencom-Rongai: context deadline exceeded", 5 * time.Second},
		{"fetch fails", time.Second, closed, 5 * time.Second, 0, closed,
			"fare for Kencom-Rongai: SACCO office closed", time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var seen context.Context
				var deadline time.Time
				var hasDeadline bool
				fetch := func(ctx context.Context, route string) (int, error) {
					seen = ctx
					deadline, hasDeadline = ctx.Deadline()
					if err := slow(ctx, tt.takes); err != nil {
						return 0, err
					}
					if tt.fails != nil {
						return 0, tt.fails
					}
					return 100, nil
				}

				start := time.Now()
				got, err := FetchFare(context.Background(), "Kencom-Rongai", tt.timeout, fetch)
				elapsed := time.Since(start)

				if got != tt.want || !errors.Is(err, tt.wantErr) {
					t.Errorf("FetchFare = %d, %v; want %d, %v", got, err, tt.want, tt.wantErr)
				}
				if tt.wantMsg != "" && (err == nil || err.Error() != tt.wantMsg) {
					t.Errorf("FetchFare error message = %v, want %q", err, tt.wantMsg)
				}
				if elapsed != tt.wantElapsed {
					t.Errorf("FetchFare returned after %v, want %v", elapsed, tt.wantElapsed)
				}
				if seen == nil {
					t.Fatal("fetch was never called")
				}
				if !hasDeadline || !deadline.Equal(start.Add(tt.timeout)) {
					t.Errorf("fetch's context has deadline %v (set: %v), want %v. Use context.WithTimeout", deadline, hasDeadline, start.Add(tt.timeout))
				}
				if seen.Err() == nil {
					t.Errorf("fetch's context is still live after FetchFare returned. Call the cancel func with defer cancel()")
				}
			})
		})
	}
}

func TestFetchFareParentCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(2*time.Second, cancel) // the customer closes the app
		fetch := func(ctx context.Context, route string) (int, error) {
			return 0, slow(ctx, time.Minute)
		}
		start := time.Now()
		_, err := FetchFare(ctx, "Odeon-Githurai", 5*time.Second, fetch)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("FetchFare with a parent cancelled at 2s = %v, want an error wrapping context.Canceled. Build the timeout on top of ctx, not context.Background()", err)
		}
		if elapsed := time.Since(start); elapsed != 2*time.Second {
			t.Errorf("FetchFare returned after %v, want 2s", elapsed)
		}
	})
}

// sacco says how a fake SACCO answers a seat check.
type sacco struct {
	after time.Duration
	free  bool
	err   error
	hangs bool // answers only when its context is done
}

func TestFirstAvailable(t *testing.T) {
	broken := errors.New("system down")
	tests := []struct {
		name         string
		saccos       map[string]sacco
		parentAfter  time.Duration // a parent deadline; 0 means none
		want         string
		wantErr      error
		wantElapsed  time.Duration
		wantStoppedN int // checks that saw their context cancelled
	}{
		{"one has a seat", map[string]sacco{
			"Super Metro": {after: 3 * time.Second, free: true},
			"Embassava":   {after: time.Second, free: false},
			"Rembo":       {hangs: true},
		}, 0, "Super Metro", nil, 3 * time.Second, 1},
		{"a failed check is a no", map[string]sacco{
			"Super Metro": {after: time.Second, err: broken},
			"Embassava":   {after: 2 * time.Second, free: true},
		}, 0, "Embassava", nil, 2 * time.Second, 0},
		{"the fastest yes wins", map[string]sacco{
			"Super Metro": {after: 4 * time.Second, free: true},
			"Embassava":   {after: 2 * time.Second, free: true},
			"Rembo":       {hangs: true},
			"Forward":     {hangs: true},
		}, 0, "Embassava", nil, 2 * time.Second, 3},
		{"everyone says no", map[string]sacco{
			"Super Metro": {after: time.Second},
			"Embassava":   {after: 2 * time.Second},
		}, 0, "", ErrNoneAvailable, 2 * time.Second, 0},
		{"parent deadline", map[string]sacco{
			"Super Metro": {hangs: true},
			"Embassava":   {after: 9 * time.Second, free: true},
		}, 5 * time.Second, "", context.DeadlineExceeded, 5 * time.Second, 2},
		{"no saccos", map[string]sacco{}, 0, "", ErrNoneAvailable, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				if tt.parentAfter > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, tt.parentAfter)
					defer cancel()
				}
				names := make([]string, 0, len(tt.saccos))
				for n := range tt.saccos {
					names = append(names, n)
				}

				var mu sync.Mutex
				asked, stopped := map[string]bool{}, 0
				check := func(ctx context.Context, name string) (bool, error) {
					mu.Lock()
					asked[name] = true
					mu.Unlock()
					s := tt.saccos[name]
					wait := s.after
					if s.hangs {
						wait = 24 * time.Hour
					}
					if err := slow(ctx, wait); err != nil {
						mu.Lock()
						stopped++
						mu.Unlock()
						return false, err
					}
					return s.free, s.err
				}

				start := time.Now()
				got, err := FirstAvailable(ctx, names, check)
				elapsed := time.Since(start)

				if got != tt.want || !errors.Is(err, tt.wantErr) {
					t.Errorf("FirstAvailable = %q, %v; want %q, %v", got, err, tt.want, tt.wantErr)
				}
				if elapsed != tt.wantElapsed {
					t.Errorf("FirstAvailable returned after %v, want %v", elapsed, tt.wantElapsed)
				}
				mu.Lock()
				defer mu.Unlock()
				if len(asked) != len(tt.saccos) {
					t.Errorf("FirstAvailable asked %d SACCOs, want all %d", len(asked), len(tt.saccos))
				}
				if stopped != tt.wantStoppedN {
					t.Errorf("%d checks had their context cancelled by the time FirstAvailable returned, want %d. Cancel the others once you have an answer, and wait for every check before returning", stopped, tt.wantStoppedN)
				}
			})
		})
	}
}
