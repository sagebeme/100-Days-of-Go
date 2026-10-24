package workerpools

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// The tests that count running goroutines use synctest.Test. Inside it,
// synctest.Wait() returns once every other goroutine is stuck waiting, so the test can
// count exactly how many jobs are running at that moment, with no sleeps and no luck.

// held is a fake job that counts itself as running until release is closed.
type held struct {
	running, peak atomic.Int32
	release       chan struct{}
}

func newHeld() *held { return &held{release: make(chan struct{})} }

func (h *held) do() {
	n := h.running.Add(1)
	for {
		p := h.peak.Load()
		if n <= p || h.peak.CompareAndSwap(p, n) {
			break
		}
	}
	<-h.release
	h.running.Add(-1)
}

func TestProcess(t *testing.T) {
	names := []string{"Akinyi", "Mutua", "Chebet"}
	got := Process(names, 2, func(n string) string { return "statement for " + n })
	want := []string{"statement for Akinyi", "statement for Mutua", "statement for Chebet"}
	if !slices.Equal(got, want) {
		t.Errorf("Process(%q, 2, render)\n got: %q\nwant: %q", names, got, want)
	}

	nums := make([]int, 100)
	for i := range nums {
		nums[i] = i
	}
	squares := Process(nums, 7, func(x int) int { return x * x })
	if len(squares) != len(nums) {
		t.Fatalf("Process(100 jobs) returned %d results, want 100", len(squares))
	}
	for i, s := range squares {
		if s != i*i {
			t.Fatalf("result %d = %d, want %d: results must come back in job order", i, s, i*i)
		}
	}

	if got := Process([]int{}, 3, func(x int) int { return x }); len(got) != 0 {
		t.Errorf("Process(no jobs) = %v, want an empty slice", got)
	}
}

func TestProcessBounded(t *testing.T) {
	tests := []struct {
		jobs, workers, wantRunning int
	}{
		{20, 3, 3},
		{20, 1, 1},
		{5, 10, 5},  // more workers than jobs: all 5 run
		{8, 0, 1},   // 0 counts as 1
		{12, -4, 1}, // so does a negative number
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d jobs %d workers", tt.jobs, tt.workers), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				h := newHeld()
				jobs := make([]int, tt.jobs)
				done := make(chan []int, 1)
				go func() {
					done <- Process(jobs, tt.workers, func(int) int { h.do(); return 1 })
				}()

				synctest.Wait() // everyone is now running or waiting for a turn
				if got := int(h.running.Load()); got != tt.wantRunning {
					t.Errorf("%d jobs were running at once, want %d", got, tt.wantRunning)
				}
				close(h.release)
				res := <-done
				if p := int(h.peak.Load()); p > tt.wantRunning {
					t.Errorf("at one point %d jobs ran at once, want at most %d", p, tt.wantRunning)
				}
				if len(res) != tt.jobs {
					t.Errorf("Process returned %d results, want %d", len(res), tt.jobs)
				}
			})
		})
	}
}

func TestFanIn(t *testing.T) {
	tests := []struct {
		name   string
		inputs [][]int
	}{
		{"three channels", [][]int{{1, 2, 3}, {10, 20}, {100}}},
		{"one channel", [][]int{{7, 8, 9}}},
		{"some are empty", [][]int{{}, {5}, {}}},
		{"no channels", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var chans []<-chan int
				var want []int
				for _, in := range tt.inputs {
					ch := make(chan int)
					go func() {
						defer close(ch)
						for _, v := range in {
							ch <- v
						}
					}()
					chans = append(chans, ch)
					want = append(want, in...)
				}

				defer func() {
					// If FanIn left any input unread, drain it so its sender can finish.
					for _, ch := range chans {
						for range ch {
						}
					}
				}()

				out := FanIn(chans...)
				if out == nil {
					t.Fatal("FanIn returned a nil channel. Make one with make(chan T)")
				}
				var got []int
				for {
					select {
					case v, ok := <-out:
						if !ok {
							slices.Sort(got)
							slices.Sort(want)
							if !slices.Equal(got, want) {
								t.Errorf("FanIn delivered %v, want %v in any order", got, want)
							}
							return
						}
						got = append(got, v)
					case <-time.After(time.Hour):
						t.Fatalf("FanIn's channel was never closed (got %v so far). Close it after every input is drained", got)
					}
				}
			})
		})
	}
}

func TestGroupWaitsForAll(t *testing.T) {
	var g Group // the zero value works
	var done atomic.Int32
	for range 10 {
		g.Go(func() error {
			done.Add(1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, want nil", err)
	}
	if n := done.Load(); n != 10 {
		t.Errorf("%d of 10 tasks had finished when Wait returned", n)
	}
}

func TestGroupFirstError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		first := errors.New("printer jammed")
		second := errors.New("out of paper")
		g, ctx := WithContext(context.Background())

		var stopped atomic.Int32
		g.Go(func() error { time.Sleep(time.Second); return first })
		g.Go(func() error { time.Sleep(2 * time.Second); return second })
		for range 3 {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					stopped.Add(1)
					return ctx.Err()
				case <-time.After(time.Hour):
					return nil
				}
			})
		}

		start := time.Now()
		err := g.Wait()
		if !errors.Is(err, first) {
			t.Errorf("Wait() = %v, want the first error, %v", err, first)
		}
		if elapsed := time.Since(start); elapsed != 2*time.Second {
			t.Errorf("Wait returned after %v, want 2s: the cancelled tasks stop at once, and the slow failure finishes at 2s", elapsed)
		}
		if n := stopped.Load(); n != 3 {
			t.Errorf("%d of 3 waiting tasks saw the context cancelled, want 3. Cancel it on the first error", n)
		}
	})
}

func TestGroupCancelsOnWait(t *testing.T) {
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return nil })
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, want nil", err)
	}
	if ctx.Err() == nil {
		t.Error("the group's context is still live after Wait returned. Cancel it in Wait")
	}
}

func TestGroupLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var g Group
		g.SetLimit(2)
		h := newHeld()

		var started atomic.Int32
		launched := make(chan struct{})
		go func() {
			for range 6 {
				g.Go(func() error { h.do(); return nil })
				started.Add(1)
			}
			close(launched)
		}()

		synctest.Wait()
		if got := h.running.Load(); got != 2 {
			t.Errorf("%d tasks were running with SetLimit(2), want 2", got)
		}
		if got := started.Load(); got != 2 {
			t.Errorf("Go had returned %d times with 2 slots full, want 2: Go must wait for a free slot", got)
		}
		close(h.release)
		<-launched
		if err := g.Wait(); err != nil {
			t.Errorf("Wait() = %v, want nil", err)
		}
		if p := h.peak.Load(); p > 2 {
			t.Errorf("at one point %d tasks ran at once, want at most 2", p)
		}
	})
}

func TestGroupIsSafe(t *testing.T) {
	// Many tasks fail at once. Exactly one error must be kept, and -race must stay quiet.
	var g Group
	var mu sync.Mutex
	sent := map[string]bool{}
	for i := range 50 {
		g.Go(func() error {
			msg := fmt.Sprintf("member %d: no phone number", i)
			mu.Lock()
			sent[msg] = true
			mu.Unlock()
			return errors.New(msg)
		})
	}
	err := g.Wait()
	if err == nil || !sent[err.Error()] {
		t.Errorf("Wait() = %v, want one of the tasks' errors", err)
	}
}
