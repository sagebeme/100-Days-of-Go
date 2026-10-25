// Package backgroundjobs is an in-process job queue for Twende's slow work, such as
// sending booking SMSes: retries with exponential backoff and jitter, deduplication by
// key, and a dead-letter list for jobs that keep failing.
package backgroundjobs

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Clock tells the queue the time. Tests pass a fake one they move by hand.
type Clock interface {
	Now() time.Time
}

// SystemClock is the real clock.
type SystemClock struct{}

// Now returns time.Now().
func (SystemClock) Now() time.Time { return time.Now() }

// Job is one piece of work. Key says what the work is about, such as
// "sms:booking-1042"; two jobs with the same key are the same work.
type Job struct {
	Key      string
	Payload  string
	Attempts int       // how many times it has failed so far
	RunAt    time.Time // not before this time
	LastErr  string    // the error from the latest failed attempt
	seq      int       // enqueue order, to break ties between equal RunAt
}

// Handler does the work for one job.
type Handler func(ctx context.Context, j Job) error

// ErrDuplicate is returned by Enqueue for a key the queue has already seen.
var ErrDuplicate = errors.New("job already queued or done")

// permanent marks an error that retrying can't fix.
type permanent struct{ err error }

func (p permanent) Error() string { return p.err.Error() }
func (p permanent) Unwrap() error { return p.err }

// Permanent wraps err so the queue sends the job straight to the dead-letter list
// instead of retrying it, for example for a phone number that doesn't exist.
func Permanent(err error) error { return permanent{err} }

// IsPermanent reports whether err, or an error it wraps, came from Permanent.
func IsPermanent(err error) bool {
	var p permanent
	return errors.As(err, &p)
}

// Backoff returns how long to wait before retry number attempt (1 for the first retry).
// The ceiling doubles each time, base, 2×base, 4×base and so on, capped at max. The
// result is half the ceiling plus a random part of the other half: ceiling/2 +
// rnd()×ceiling/2, where rnd returns a number in [0, 1). The random part (jitter)
// stops many failed jobs from all retrying at the same moment.
//
//	Backoff(1, time.Second, time.Minute, func() float64 { return 0 })   → 500ms
//	Backoff(3, time.Second, time.Minute, func() float64 { return 0.5 }) → 3s   (ceiling 4s)
func Backoff(attempt int, base, max time.Duration, rnd func() float64) time.Duration {
	// Your code here
	return 0
}

// Options configures a Queue.
type Options struct {
	MaxAttempts int           // failures before a job is dead-lettered
	BaseDelay   time.Duration // first backoff ceiling
	MaxDelay    time.Duration // largest backoff ceiling
	Rand        func() float64
}

// Queue holds jobs until they're due and runs them with its handler.
// It's safe to use from several goroutines.
type Queue struct {
	mu      sync.Mutex
	clock   Clock
	handler Handler
	opts    Options
	pending []Job
	dead    []Job
	seen    map[string]bool // every key ever accepted
	nextSeq int
}

// NewQueue returns an empty queue that runs jobs with handler.
func NewQueue(clock Clock, handler Handler, opts Options) *Queue {
	return &Queue{clock: clock, handler: handler, opts: opts, seen: map[string]bool{}}
}

// Enqueue adds a job to run now. It returns ErrDuplicate, and adds nothing, if a job
// with the same key was enqueued before, whether it's still waiting, done or dead.
func (q *Queue) Enqueue(key, payload string) error {
	// Your code here
	return nil
}

// RunDue runs every job whose RunAt is not after the clock's time, earliest first (and
// in enqueue order for equal times), and returns how many it ran. A job that fails is
// retried after Backoff(its new Attempts count); once it has failed MaxAttempts times,
// or fails with a Permanent error, it moves to the dead-letter list. Jobs that become
// due during the call wait for the next one. The handler runs without the lock held.
func (q *Queue) RunDue(ctx context.Context) int {
	// Your code here
	return 0
}

// Run calls RunDue every time tick delivers, until ctx is cancelled.
func (q *Queue) Run(ctx context.Context, tick <-chan time.Time) {
	// Your code here
}

// Pending returns a copy of the jobs still waiting, earliest RunAt first.
func (q *Queue) Pending() []Job {
	// Your code here
	return nil
}

// DeadLetters returns a copy of the jobs that gave up, in the order they died.
func (q *Queue) DeadLetters() []Job {
	// Your code here
	return nil
}

// byRunAt orders jobs by RunAt, then by the order they were enqueued.
func byRunAt(a, b Job) int {
	if c := a.RunAt.Compare(b.RunAt); c != 0 {
		return c
	}
	return a.seq - b.seq
}

// Idempotent wraps h so that a job key that already succeeded is not run again. Queues
// that deliver "at least once" can hand a worker the same job twice, for example after
// a crash between doing the work and saying so; this makes the second time harmless.
// Failed runs are not remembered, so they can be retried. Two copies running at the
// same moment can both get through; a real system closes that gap with a unique
// constraint on the key in the database.
func Idempotent(h Handler) Handler {
	// Your code here
	return h
}
