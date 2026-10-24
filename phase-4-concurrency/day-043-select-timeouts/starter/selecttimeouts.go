// Package selecttimeouts waits on several channels at once: the fastest airtime quote
// wins, a full matatu says so instead of making you wait, and two ticket counters feed
// one queue.
package selecttimeouts

import (
	"errors"
	"time"
)

// Quote is one network's price for a bundle, in shillings.
type Quote struct {
	Network string
	Price   int
}

// ErrTimeout means no quote arrived before the timeout.
var ErrTimeout = errors.New("no quote in time")

// ErrNoQuotes means both channels were closed without sending a quote.
var ErrNoQuotes = errors.New("no network sent a quote")

// Fastest returns the first quote to arrive on a or b. If timeout passes first it
// returns ErrTimeout. A closed channel has nothing to send, so Fastest stops listening to
// it; if both are closed it returns ErrNoQuotes.
//
//	b sends after 2s, a never does, timeout 5s  → b's quote, nil       (after 2s)
//	neither sends, timeout 5s                   → Quote{}, ErrTimeout  (after 5s)
func Fastest(a, b <-chan Quote, timeout time.Duration) (Quote, error) {
	// Your code here. Follow the steps in the README.
	return Quote{}, nil
}

// TryBook puts name on seats if there is a free seat, and reports whether it did. seats
// is a buffered channel with one slot per seat. TryBook never waits: a full matatu gives
// false at once.
//
//	TryBook(seats, "Achieng") → true    (a seat was free)
//	TryBook(seats, "Kamau")   → false   (the matatu is full)
func TryBook(seats chan<- string, name string) bool {
	// Your code here.
	return false
}

// Merge returns a channel that delivers every value from a and b as it arrives, and is
// closed once both a and b are closed.
func Merge(a, b <-chan string) <-chan string {
	// Your code here.
	return nil
}
