// Package contextcancel looks up matatu fares and free seats without waiting forever:
// every slow call takes a context, and stops when that context is cancelled or its
// deadline passes.
package contextcancel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNoneAvailable means every SACCO answered, and none had a seat.
var ErrNoneAvailable = errors.New("no SACCO has a free seat")

// Sleep waits for d, or until ctx is done, whichever comes first. It returns nil if the
// full d passed, and ctx.Err() if ctx was done first.
//
//	Sleep(ctx, 5*time.Second)                       → nil                        after 5s
//	Sleep(ctx, 5*time.Second), ctx cancelled at 2s  → context.Canceled           after 2s
func Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// FetchFare calls fetch for route with a context that times out after timeout, and
// cancels that context before it returns. If fetch fails, the error says which route it
// was for and wraps fetch's error, so errors.Is still finds context.DeadlineExceeded.
//
//	FetchFare(ctx, "Kencom-Rongai", 5*time.Second, fetch)   fetch answers in 1s   → 100, nil
//	FetchFare(ctx, "Kencom-Rongai", 5*time.Second, fetch)   fetch takes 10s       → 0, "fare for Kencom-Rongai: context deadline exceeded"
func FetchFare(ctx context.Context, route string, timeout time.Duration, fetch func(ctx context.Context, route string) (int, error)) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // frees the timer even when fetch answers early
	fare, err := fetch(ctx, route)
	if err != nil {
		return 0, fmt.Errorf("fare for %s: %w", route, err)
	}
	return fare, nil
}

// FirstAvailable asks every SACCO in saccos at the same time whether it has a free seat,
// and returns the first one that says yes. As soon as it has an answer, it cancels the
// context it gave the other checks, so they can stop early. A check that fails counts as
// a no. FirstAvailable returns only after every check has returned.
//
// If ctx is done before any SACCO says yes, it returns ctx.Err(). If every SACCO says no,
// it returns ErrNoneAvailable.
func FirstAvailable(ctx context.Context, saccos []string, check func(ctx context.Context, sacco string) (bool, error)) (string, error) {
	checkCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Buffered, so no check ever waits to report, even after we've stopped listening.
	answers := make(chan string, len(saccos))
	var wg sync.WaitGroup
	for _, s := range saccos {
		wg.Go(func() {
			ok, err := check(checkCtx, s)
			if err != nil || !ok {
				s = ""
			}
			answers <- s
		})
	}

	winner := ""
	for range saccos {
		if s := <-answers; s != "" && winner == "" {
			winner = s
			cancel() // tell the others to stop
		}
	}
	wg.Wait()

	switch {
	case winner != "":
		return winner, nil
	case ctx.Err() != nil:
		return "", ctx.Err()
	default:
		return "", ErrNoneAvailable
	}
}
