// Package workerpools renders a chama's monthly statements with a fixed number of
// workers, merges results from many channels into one, and runs a group of tasks that
// stops at the first error, using only the standard library.
package workerpools

import "context"

// Process runs work on every job using at most workers goroutines at a time, and
// returns the results in the same order as jobs. A workers value below 1 counts as 1.
//
//	Process([]string{"Akinyi", "Mutua", "Chebet"}, 2, render)
//	→ []string{"statement for Akinyi", "statement for Mutua", "statement for Chebet"}
//	   (never more than 2 renders at once)
func Process[J, R any](jobs []J, workers int, work func(J) R) []R {
	// Your code here. Follow the steps in the README.
	return nil
}

// FanIn returns a channel that delivers every value from every channel in chans, and is
// closed once all of them are closed.
func FanIn[T any](chans ...<-chan T) <-chan T {
	// Your code here.
	return nil
}

// Group runs tasks in goroutines and collects the first error, like errgroup.Group from
// golang.org/x/sync. The zero value is ready to use, with no limit and no context.
type Group struct {
	// Your code here.
}

// WithContext returns a new Group and a context made from ctx. The context is cancelled
// the first time a task returns an error, or when Wait returns, whichever comes first.
func WithContext(ctx context.Context) (*Group, context.Context) {
	// Your code here.
	return &Group{}, ctx
}

// SetLimit makes at most n tasks run at once: Go waits for a free slot before starting
// another. n below 1 means no limit. Call it before the first Go.
func (g *Group) SetLimit(n int) {
	// Your code here.
}

// Go runs f in a new goroutine. If the group has a limit and it's reached, Go waits
// until a running task finishes.
func (g *Group) Go(f func() error) {
	// Your code here.
}

// Wait waits for every task started with Go, then returns the first error any of them
// returned, or nil.
func (g *Group) Wait() error {
	// Your code here.
	return nil
}
