package mutexes

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// Run these with the race detector too:
//
//	go test -race ./phase-4-concurrency/day-045-mutexes/starter

func TestTotalContributions(t *testing.T) {
	many := make([]int, 2000)
	for i := range many {
		many[i] = 50
	}
	tests := []struct {
		name    string
		amounts []int
		want    int
	}{
		{"three members", []int{500, 1000, 250}, 1750},
		{"nobody paid", nil, 0},
		{"a big chama", many, 100000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalContributions(tt.amounts); got != tt.want {
				t.Errorf("TotalContributions(%d payments) = %d, want %d", len(tt.amounts), got, tt.want)
			}
		})
	}
}

func TestLedger(t *testing.T) {
	l := NewLedger()
	steps := []struct {
		op      string
		member  string
		amount  int
		wantErr error
		want    int // the member's balance afterwards
	}{
		{"deposit", "Akinyi", 500, nil, 500},
		{"withdraw", "Akinyi", 300, nil, 200},
		{"withdraw", "Akinyi", 300, ErrInsufficientFunds, 200},
		{"withdraw", "Akinyi", 200, nil, 0},
		{"deposit", "Mutua", 1000, nil, 1000},
		{"withdraw", "Chebet", 1, ErrInsufficientFunds, 0},
	}
	for _, s := range steps {
		var err error
		if s.op == "deposit" {
			l.Deposit(s.member, s.amount)
		} else {
			err = l.Withdraw(s.member, s.amount)
		}
		if !errors.Is(err, s.wantErr) {
			t.Errorf("%s(%q, %d) error = %v, want %v", s.op, s.member, s.amount, err, s.wantErr)
		}
		if got := l.Balance(s.member); got != s.want {
			t.Errorf("after %s(%q, %d): Balance = %d, want %d", s.op, s.member, s.amount, got, s.want)
		}
	}
	if got := l.Total(); got != 1000 {
		t.Errorf("Total() = %d, want 1000", got)
	}
}

func TestLedgerConcurrentDeposits(t *testing.T) {
	l := NewLedger()
	var wg sync.WaitGroup
	for i := range 100 {
		member := fmt.Sprintf("member-%d", i%5)
		wg.Go(func() {
			for range 100 {
				l.Deposit(member, 10)
			}
		})
	}
	wg.Wait()
	for i := range 5 {
		member := fmt.Sprintf("member-%d", i)
		if got := l.Balance(member); got != 20000 {
			t.Errorf("Balance(%q) = %d after 2000 deposits of 10, want 20000", member, got)
		}
	}
	if got := l.Total(); got != 100000 {
		t.Errorf("Total() = %d, want 100000", got)
	}
}

func TestLedgerConcurrentWithdrawals(t *testing.T) {
	// 50 people try to take Ksh100 from a balance of Ksh1000 at the same moment.
	// Exactly 10 must succeed, and the balance must never go below zero.
	l := NewLedger()
	l.Deposit("chama", 1000)

	var wg sync.WaitGroup
	results := make([]error, 50)
	for i := range results {
		wg.Go(func() {
			results[i] = l.Withdraw("chama", 100)
		})
	}
	wg.Wait()

	ok, refused := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrInsufficientFunds):
			refused++
		default:
			t.Errorf("Withdraw returned %v, want nil or ErrInsufficientFunds", err)
		}
	}
	if ok != 10 || refused != 40 {
		t.Errorf("%d withdrawals succeeded and %d were refused, want 10 and 40", ok, refused)
	}
	if got := l.Balance("chama"); got != 0 {
		t.Errorf("Balance = %d after the rush, want 0", got)
	}
}

func TestLedgerReadsWhileWriting(t *testing.T) {
	// Balance and Total read the map while deposits write it. Without the lock this is
	// "fatal error: concurrent map read and map write", or a race report under -race.
	l := NewLedger()
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for i := range 200 {
				l.Deposit(fmt.Sprintf("m%d", i%20), 1)
			}
		})
		wg.Go(func() {
			for range 200 {
				_ = l.Balance("m1")
				_ = l.Total()
			}
		})
	}
	wg.Wait()
	if got := l.Total(); got != 800 {
		t.Errorf("Total() = %d, want 800", got)
	}
}
