// Package pricewatcher checks the price of every item at many shops at once, without
// hammering any of them: a bounded worker pool, a timeout per request, a global rate
// limit, and retries with backoff when a shop's server has a bad moment.
package pricewatcher

import (
	"context"
	"net/http"
	"time"
)

// Shop is one shop's price service. GET URL + "/prices" answers with a JSON list:
//
//	[{"item": "unga 2kg", "price": 230}, {"item": "sukari 1kg", "price": 185}]
type Shop struct {
	Name string
	URL  string
}

// Offer is the price of an item at one shop, in shillings.
type Offer struct {
	Shop  string
	Price int
}

// Summary is what Watch found.
type Summary struct {
	Cheapest map[string]Offer // item → the cheapest offer for it
	Failed   []string         // names of the shops that couldn't be read, sorted
}

// Limiter paces requests. *rate.Limiter from golang.org/x/time/rate satisfies it.
type Limiter interface {
	Wait(ctx context.Context) error
}

// Config controls how Watch talks to the shops.
type Config struct {
	Workers int           // the most requests in flight at once; below 1 counts as 1
	Timeout time.Duration // the limit for each request attempt; 0 means none
	Retries int           // extra attempts after a 5xx answer
	Backoff time.Duration // the wait before the first retry; it doubles for each one after
	Limiter Limiter       // waited on before every attempt; nil means no limit

	// Sleep waits between retries. Nil means a real wait that stops early if ctx is done.
	// Tests pass their own, so they don't wait for real.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Watch fetches the prices of every shop and returns the cheapest offer for each item.
// A shop that fails (a timeout, a non-200 answer, bad JSON, or still a 5xx after every
// retry) is listed in Summary.Failed and doesn't stop the others. Entries with no item
// name or a price below 1 are ignored. When two shops have the same lowest price, the
// one whose name sorts first wins. Watch returns an error only if ctx is done.
func Watch(ctx context.Context, client *http.Client, shops []Shop, cfg Config) (Summary, error) {
	// Your code here. The README lists what Watch must do.
	return Summary{}, nil
}

// Report formats s for the terminal: a heading, then one line per item in alphabetical
// order, then the failed shops if there are any.
//
//	Cheapest prices:
//	sukari 1kg: Ksh180 at Duka la Kona
//	unga 2kg: Ksh230 at Mama Njeri Shop
//	Failed: Gikomba Wholesale
func Report(s Summary) string {
	// Your code here.
	return ""
}
