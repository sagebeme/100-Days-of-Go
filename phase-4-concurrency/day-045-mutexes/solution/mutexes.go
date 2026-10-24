// Package mutexes keeps a chama's books while M-Pesa payments arrive from many goroutines
// at once. Shared state is guarded by a sync.Mutex, and go test -race proves it.
package mutexes

import (
	"errors"
	"sync"
)

// ErrInsufficientFunds means a withdrawal is bigger than the member's balance.
var ErrInsufficientFunds = errors.New("insufficient funds")

// TotalContributions adds up every member's contribution, one goroutine per payment.
func TotalContributions(amounts []int) int {
	var mu sync.Mutex
	total := 0
	var wg sync.WaitGroup
	for _, a := range amounts {
		wg.Go(func() {
			mu.Lock()
			total += a
			mu.Unlock()
		})
	}
	wg.Wait()
	return total
}

// Ledger holds each member's balance in shillings. Every method is safe to call from many
// goroutines at the same time. Make one with NewLedger.
type Ledger struct {
	mu       sync.Mutex
	balances map[string]int
}

// NewLedger returns an empty ledger, ready to use.
func NewLedger() *Ledger {
	return &Ledger{balances: map[string]int{}}
}

// Deposit adds amount to member's balance.
func (l *Ledger) Deposit(member string, amount int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.balances[member] += amount
}

// Withdraw takes amount from member's balance. If the balance is smaller than amount it
// changes nothing and returns ErrInsufficientFunds.
//
//	Deposit("Akinyi", 500); Withdraw("Akinyi", 300) → nil, balance 200
//	Withdraw("Akinyi", 300)                         → ErrInsufficientFunds, balance still 200
func (l *Ledger) Withdraw(member string, amount int) error {
	// The check and the change happen under one lock. Unlocking between them would let
	// two withdrawals both see the same balance and both take it.
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.balances[member] < amount {
		return ErrInsufficientFunds
	}
	l.balances[member] -= amount
	return nil
}

// Balance returns member's balance; 0 for a member the ledger hasn't seen.
func (l *Ledger) Balance(member string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balances[member]
}

// Total returns the sum of every member's balance.
func (l *Ledger) Total() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	total := 0
	for _, b := range l.balances {
		total += b
	}
	return total
}
