package caching

import "sync"

// call is one piece of work in progress and everyone waiting for it.
type call[T any] struct {
	done    chan struct{}
	val     T
	err     error
	callers int
}

// Group makes sure that, for each key, only one fn runs at a time. Callers who arrive
// while it's running wait and share its result instead of starting their own. It's a
// small version of golang.org/x/sync/singleflight. The zero value is ready to use.
type Group[T any] struct {
	mu    sync.Mutex
	calls map[string]*call[T]
}

// Do runs fn for key, or waits for the run already in progress, and returns its result.
// Once the run finishes, the next Do for key runs fn again.
func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error) {
	// Your code here
	var zero T
	return zero, nil
}

// Callers returns how many callers are sharing the run in progress for key: 0 if none is
// running. It's handy for metrics and for tests.
func (g *Group[T]) Callers(key string) int {
	// Your code here
	return 0
}
