// Package workerpools renders a chama's monthly statements with a fixed number of
// workers, merges results from many channels into one, and runs a group of tasks that
// stops at the first error, using only the standard library.
package workerpools

import (
	"context"
	"sync"
)

// Process runs work on every job using at most workers goroutines at a time, and
// returns the results in the same order as jobs. A workers value below 1 counts as 1.
//
//	Process([]string{"Akinyi", "Mutua", "Chebet"}, 2, render)
//	→ []string{"statement for Akinyi", "statement for Mutua", "statement for Chebet"}
//	   (never more than 2 renders at once)
func Process[J, R any](jobs []J, workers int, work func(J) R) []R {
	results := make([]R, len(jobs))
	next := make(chan int) // job numbers, handed out one at a time

	var wg sync.WaitGroup
	for range max(workers, 1) {
		wg.Go(func() {
			for i := range next {
				results[i] = work(jobs[i])
			}
		})
	}
	for i := range jobs {
		next <- i
	}
	close(next) // the workers' range loops end once every job is taken
	wg.Wait()
	return results
}

// FanIn returns a channel that delivers every value from every channel in chans, and is
// closed once all of them are closed.
func FanIn[T any](chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Go(func() {
			for v := range ch {
				out <- v
			}
		})
	}
	// Close out only after every forwarding goroutine is done. Closing it any earlier
	// would make one of them send on a closed channel.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Group runs tasks in goroutines and collects the first error, like errgroup.Group from
// golang.org/x/sync. The zero value is ready to use, with no limit and no context.
type Group struct {
	wg     sync.WaitGroup
	slots  chan struct{} // nil means no limit
	once   sync.Once
	err    error
	cancel context.CancelFunc
}

// WithContext returns a new Group and a context made from ctx. The context is cancelled
// the first time a task returns an error, or when Wait returns, whichever comes first.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// SetLimit makes at most n tasks run at once: Go waits for a free slot before starting
// another. n below 1 means no limit. Call it before the first Go.
func (g *Group) SetLimit(n int) {
	if n < 1 {
		g.slots = nil
		return
	}
	g.slots = make(chan struct{}, n)
}

// Go runs f in a new goroutine. If the group has a limit and it's reached, Go waits
// until a running task finishes.
func (g *Group) Go(f func() error) {
	if g.slots != nil {
		g.slots <- struct{}{} // waits while every slot is taken
	}
	g.wg.Go(func() {
		defer func() {
			if g.slots != nil {
				<-g.slots
			}
		}()
		if err := f(); err != nil {
			g.once.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	})
}

// Wait waits for every task started with Go, then returns the first error any of them
// returned, or nil.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
