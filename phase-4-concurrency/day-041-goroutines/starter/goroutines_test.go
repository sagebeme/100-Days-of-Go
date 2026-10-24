package goroutines

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// gate holds every caller until n callers have arrived. If they don't all arrive within
// two seconds, it gives up and lets everyone through, and apart reports true. Calls made
// one after another can never all be inside the gate at once, so this tells concurrent
// code apart from sequential code without timing anything.
type gate struct {
	arrived sync.WaitGroup
	all     chan struct{}
	expired chan struct{}
	apart   atomic.Bool
}

func newGate(n int) *gate {
	g := &gate{all: make(chan struct{}), expired: make(chan struct{})}
	g.arrived.Add(n)
	go func() {
		g.arrived.Wait()
		close(g.all)
	}()
	time.AfterFunc(2*time.Second, func() { close(g.expired) })
	return g
}

func (g *gate) arrive() {
	g.arrived.Done()
	select {
	case <-g.all:
	case <-g.expired:
		g.apart.Store(true)
	}
}

func TestSendAll(t *testing.T) {
	tests := []struct {
		name    string
		numbers []string
	}{
		{"two members", []string{"0712345678", "0723456789"}},
		{"one member", []string{"0700111222"}},
		{"a whole chama", []string{
			"0711000001", "0711000002", "0711000003", "0711000004", "0711000005",
			"0711000006", "0711000007", "0711000008", "0711000009", "0711000010",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGate(len(tt.numbers))
			var calls atomic.Int32
			send := func(number string) string {
				calls.Add(1)
				g.arrive()
				return number + ": sent"
			}

			got := SendAll(tt.numbers, send)

			want := make([]string, len(tt.numbers))
			for i, n := range tt.numbers {
				want[i] = n + ": sent"
			}
			if !slices.Equal(got, want) {
				t.Errorf("SendAll(%q)\n got: %q\nwant: %q", tt.numbers, got, want)
			}
			if c := int(calls.Load()); c != len(tt.numbers) {
				t.Errorf("send was called %d times, want %d (once per number)", c, len(tt.numbers))
			}
			if g.apart.Load() {
				t.Errorf("the sends never all ran at the same time: SendAll seems to send one at a time")
			}
		})
	}
}

func TestSendAllKeepsOrder(t *testing.T) {
	// The first number's send finishes last. Its result must still come first.
	numbers := []string{"0711000001", "0711000002", "0711000003"}
	others := make(chan struct{})
	var done sync.WaitGroup
	done.Add(len(numbers) - 1)
	go func() {
		done.Wait()
		close(others)
	}()
	send := func(number string) string {
		if number == numbers[0] {
			select {
			case <-others:
			case <-time.After(2 * time.Second):
			}
			return number + ": slow"
		}
		defer done.Done()
		return number + ": fast"
	}
	got := SendAll(numbers, send)
	want := []string{"0711000001: slow", "0711000002: fast", "0711000003: fast"}
	if !slices.Equal(got, want) {
		t.Errorf("SendAll(%q)\n got: %q\nwant: %q", numbers, got, want)
	}
}

func TestSendAllNoNumbers(t *testing.T) {
	got := SendAll(nil, func(string) string { return "should not be called" })
	if len(got) != 0 {
		t.Errorf("SendAll(nil) = %q, want an empty slice", got)
	}
}

func TestSumBalances(t *testing.T) {
	tests := []struct {
		name     string
		balances map[string]int
		want     int
	}{
		{"three accounts", map[string]int{"mpesa": 12500, "bank": 60000, "sacco": 13000}, 85500},
		{"one account", map[string]int{"mpesa": 450}, 450},
		{"no accounts", map[string]int{}, 0},
		{"many accounts", func() map[string]int {
			m := map[string]int{}
			for i := 1; i <= 20; i++ {
				m[fmt.Sprintf("member-%02d", i)] = i * 100
			}
			return m
		}(), 21000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := make([]string, 0, len(tt.balances))
			for a := range tt.balances {
				accounts = append(accounts, a)
			}
			slices.Sort(accounts)
			g := newGate(len(accounts))
			balance := func(account string) int {
				g.arrive()
				return tt.balances[account]
			}
			got := SumBalances(accounts, balance)
			if got != tt.want {
				t.Errorf("SumBalances(%q) = %d, want %d", accounts, got, tt.want)
			}
			if g.apart.Load() {
				t.Errorf("the balance calls never all ran at the same time: SumBalances seems to ask one at a time")
			}
		})
	}
}
