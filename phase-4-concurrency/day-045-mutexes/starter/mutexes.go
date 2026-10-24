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
//
// It has a data race. Run go test -race on this package to see it, then fix it with a
// sync.Mutex. Keep the goroutines: the point is to make them safe, not to remove them.
func TotalContributions(amounts []int) int {
	total := 0
	var wg sync.WaitGroup
	for _, a := range amounts {
		wg.Go(func() {
			total += a
		})
	}
	wg.Wait()
	return total
}

// Ledger holds each member's balance in shillings. Every method is safe to call from many
// goroutines at the same time. Make one with NewLedger.
type Ledger struct {
	// Your code here: a sync.Mutex, and a map from member to balance.
}

// NewLedger returns an empty ledger, ready to use.
func NewLedger() *Ledger {
	// Your code here.
	return &Ledger{}
}

// Deposit adds amount to member's balance.
func (l *Ledger) Deposit(member string, amount int) {
	// Your code here.
}

// Withdraw takes amount from member's balance. If the balance is smaller than amount it
// changes nothing and returns ErrInsufficientFunds.
//
//	Deposit("Akinyi", 500); Withdraw("Akinyi", 300) → nil, balance 200
//	Withdraw("Akinyi", 300)                         → ErrInsufficientFunds, balance still 200
func (l *Ledger) Withdraw(member string, amount int) error {
	// Your code here.
	return nil
}

// Balance returns member's balance; 0 for a member the ledger hasn't seen.
func (l *Ledger) Balance(member string) int {
	// Your code here.
	return 0
}

// Total returns the sum of every member's balance.
func (l *Ledger) Total() int {
	// Your code here.
	return 0
}
