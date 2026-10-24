package selecttimeouts

import (
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

// source describes what a fake network does in a test.
type source struct {
	after  time.Duration // when it sends its quote
	quote  Quote
	never  bool // it never sends
	closed bool // its channel is closed straight away
}

func TestFastest(t *testing.T) {
	saf := Quote{"Safaricom", 99}
	air := Quote{"Airtel", 95}
	tests := []struct {
		name        string
		a, b        source
		timeout     time.Duration
		want        Quote
		wantErr     error
		wantElapsed time.Duration
	}{
		{"b answers first", source{never: true}, source{after: 2 * time.Second, quote: air}, 5 * time.Second, air, nil, 2 * time.Second},
		{"a answers first", source{after: time.Second, quote: saf}, source{after: 3 * time.Second, quote: air}, 5 * time.Second, saf, nil, time.Second},
		{"nobody answers", source{never: true}, source{never: true}, 5 * time.Second, Quote{}, ErrTimeout, 5 * time.Second},
		{"answer comes too late", source{after: 6 * time.Second, quote: saf}, source{never: true}, 5 * time.Second, Quote{}, ErrTimeout, 5 * time.Second},
		{"a closed, b answers", source{closed: true}, source{after: 4 * time.Second, quote: air}, 5 * time.Second, air, nil, 4 * time.Second},
		{"a closed, b too slow", source{closed: true}, source{after: 9 * time.Second, quote: air}, 5 * time.Second, Quote{}, ErrTimeout, 5 * time.Second},
		{"both closed", source{closed: true}, source{closed: true}, 5 * time.Second, Quote{}, ErrNoQuotes, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inside synctest.Test, time is fake: it jumps forward whenever every goroutine
			// is waiting, so a 5 second timeout takes no real time and is exactly 5s long.
			synctest.Test(t, func(t *testing.T) {
				stop := make(chan struct{})
				a, b := make(chan Quote), make(chan Quote)
				for _, s := range []struct {
					src source
					ch  chan Quote
				}{{tt.a, a}, {tt.b, b}} {
					switch {
					case s.src.closed:
						close(s.ch)
					case !s.src.never:
						go func() {
							select {
							case <-time.After(s.src.after):
							case <-stop:
								return
							}
							select {
							case s.ch <- s.src.quote:
							case <-stop:
							}
						}()
					}
				}
				defer func() {
					// Let the fake networks go home, then close what's still open so a stuck
					// Fastest can return too.
					close(stop)
					synctest.Wait()
					if !tt.a.closed {
						close(a)
					}
					if !tt.b.closed {
						close(b)
					}
				}()

				type result struct {
					q   Quote
					err error
				}
				done := make(chan result, 1)
				start := time.Now()
				go func() {
					q, err := Fastest(a, b, tt.timeout)
					done <- result{q, err}
				}()
				var r result
				select {
				case r = <-done:
				case <-time.After(time.Hour):
					t.Fatalf("Fastest(%v timeout) was still waiting an hour later. Add a case for the timeout to your select", tt.timeout)
				}
				elapsed := time.Since(start)

				if !errors.Is(r.err, tt.wantErr) {
					t.Errorf("Fastest error = %v, want %v", r.err, tt.wantErr)
				}
				if r.q != tt.want {
					t.Errorf("Fastest quote = %v, want %v", r.q, tt.want)
				}
				if elapsed != tt.wantElapsed {
					t.Errorf("Fastest returned after %v, want %v", elapsed, tt.wantElapsed)
				}
			})
		})
	}
}

func TestTryBook(t *testing.T) {
	seats := make(chan string, 2)
	steps := []struct {
		name string
		want bool
	}{
		{"Achieng", true},
		{"Kamau", true},
		{"Njeri", false},
		{"Otieno", false},
	}
	for _, s := range steps {
		got := make(chan bool, 1)
		go func() { got <- TryBook(seats, s.name) }()
		select {
		case ok := <-got:
			if ok != s.want {
				t.Errorf("TryBook(%q) with %d of %d seats taken = %v, want %v", s.name, len(seats), cap(seats), ok, s.want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("TryBook(%q) on a full matatu waited instead of returning false. Add a default case", s.name)
		}
	}
	close(seats)
	var booked []string
	for name := range seats {
		booked = append(booked, name)
	}
	if want := []string{"Achieng", "Kamau"}; !slices.Equal(booked, want) {
		t.Errorf("the seats hold %q, want %q", booked, want)
	}
}

func TestTryBookNoSeats(t *testing.T) {
	seats := make(chan string) // unbuffered: no seats at all, and nobody receiving
	got := make(chan bool, 1)
	go func() { got <- TryBook(seats, "Wafula") }()
	select {
	case ok := <-got:
		if ok {
			t.Errorf("TryBook on a matatu with no seats = true, want false")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("TryBook waited instead of returning false. Add a default case")
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
	}{
		{"both counters", []string{"A1", "A2", "A3"}, []string{"B1", "B2"}},
		{"a is empty", nil, []string{"B1", "B2"}},
		{"b is empty", []string{"A1"}, nil},
		{"both empty", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := make(chan string), make(chan string)
			go func() {
				defer close(a)
				for _, v := range tt.a {
					a <- v
				}
			}()
			go func() {
				defer close(b)
				for _, v := range tt.b {
					b <- v
				}
			}()

			ch := make(chan (<-chan string), 1)
			go func() { ch <- Merge(a, b) }()
			var out <-chan string
			select {
			case out = <-ch:
			case <-time.After(2 * time.Second):
				t.Fatal("Merge didn't return. Do the selecting in a goroutine and return the channel at once")
			}
			if out == nil {
				t.Fatal("Merge returned a nil channel. Make one with make(chan string)")
			}

			want := len(tt.a) + len(tt.b)
			var got []string
			timeout := time.After(2 * time.Second)
		loop:
			for {
				select {
				case v, ok := <-out:
					if !ok {
						break loop
					}
					got = append(got, v)
					if len(got) > want {
						t.Fatalf("Merge sent more values than went in: %q. A closed channel gives \"\" forever; set it to nil to stop selecting on it", got[:min(len(got), 10)])
					}
				case <-timeout:
					t.Fatalf("Merge's channel was never closed (received %q so far). Close it once both inputs are closed", got)
				}
			}

			// Values from one input keep their order; how the two interleave is up to the scheduler.
			var fromA, fromB []string
			for _, v := range got {
				if v[0] == 'A' {
					fromA = append(fromA, v)
				} else {
					fromB = append(fromB, v)
				}
			}
			if !slices.Equal(fromA, tt.a) || !slices.Equal(fromB, tt.b) {
				t.Errorf("Merge delivered %q; want every value of %q and %q, each in its own order", got, tt.a, tt.b)
			}
		})
	}
}
