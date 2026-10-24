// Package ratelimiting keeps a bulk-SMS sender under its gateway's limit: a ticker paces
// the sends, a token bucket allows short bursts, and golang.org/x/time/rate does the
// same job per client.
package ratelimiting

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Throttle calls send for each item, in order, waiting for one value from tick before
// each send. It returns nil once every item is sent, or ctx.Err() if ctx is done first.
// In real code tick is a time.Ticker's C; tests pass a channel they control.
//
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//	Throttle(ctx, numbers, ticker.C, send)   // two sends a second
func Throttle(ctx context.Context, items []string, tick <-chan time.Time, send func(item string)) error {
	// Your code here. Follow the steps in the README.
	return nil
}

// TokenBucket allows bursts of up to capacity events, then one more event per every.
// It starts full. Tokens are added at fixed steps of every after the start (or after
// the bucket was last full), and a full bucket can't hold more. It is safe to use from
// many goroutines.
type TokenBucket struct {
	// Your code here.
}

// NewTokenBucket returns a full bucket that holds capacity tokens and gains one every
// every, starting from now.
func NewTokenBucket(capacity int, every time.Duration, now time.Time) *TokenBucket {
	// Your code here.
	return &TokenBucket{}
}

// Allow reports whether an event may happen at now, and if so uses up one token. Times
// passed to Allow never go backwards.
//
//	b := NewTokenBucket(2, time.Second, t0)
//	b.Allow(t0)                           → true   (1 token left)
//	b.Allow(t0)                           → true   (0 left)
//	b.Allow(t0.Add(500*time.Millisecond)) → false  (the next token comes at t0+1s)
//	b.Allow(t0.Add(time.Second))          → true
func (b *TokenBucket) Allow(now time.Time) bool {
	// Your code here.
	return false
}

// ClientLimiter gives every client its own rate.Limiter, so one noisy client can't use
// up everyone's allowance. It is safe to use from many goroutines.
type ClientLimiter struct {
	// Your code here.
}

// NewClientLimiter returns a ClientLimiter whose clients may each make events at rate r
// with bursts of up to burst.
func NewClientLimiter(r rate.Limit, burst int) *ClientLimiter {
	// Your code here.
	return &ClientLimiter{}
}

// Allow reports whether client may make one event at now. A client seen for the first
// time starts with a full burst.
//
//	l := NewClientLimiter(rate.Every(time.Second), 2)
//	l.Allow("shop-a", t0), l.Allow("shop-a", t0), l.Allow("shop-a", t0) → true, true, false
//	l.Allow("shop-b", t0)                                              → true
func (l *ClientLimiter) Allow(client string, now time.Time) bool {
	// Your code here.
	return false
}
