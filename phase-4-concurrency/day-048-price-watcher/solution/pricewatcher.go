// Package pricewatcher checks the price of every item at many shops at once, without
// hammering any of them: a bounded worker pool, a timeout per request, a global rate
// limit, and retries with backoff when a shop's server has a bad moment.
package pricewatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
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

// maxBody is the most Watch reads from one shop's answer.
const maxBody = 1 << 20

// errServer marks a 5xx answer: the shop's fault, so worth trying again.
var errServer = errors.New("server error")

type entry struct {
	Item  string `json:"item"`
	Price int    `json:"price"`
}

// Watch fetches the prices of every shop and returns the cheapest offer for each item.
// A shop that fails (a timeout, a non-200 answer, bad JSON, or still a 5xx after every
// retry) is listed in Summary.Failed and doesn't stop the others. Entries with no item
// name or a price below 1 are ignored. When two shops have the same lowest price, the
// one whose name sorts first wins. Watch returns an error only if ctx is done.
func Watch(ctx context.Context, client *http.Client, shops []Shop, cfg Config) (Summary, error) {
	if cfg.Sleep == nil {
		cfg.Sleep = sleep
	}
	results := make([][]entry, len(shops))
	errs := make([]error, len(shops))

	next := make(chan int)
	var wg sync.WaitGroup
	for range max(cfg.Workers, 1) {
		wg.Go(func() {
			for i := range next {
				results[i], errs[i] = fetch(ctx, client, shops[i].URL, cfg)
			}
		})
	}
	for i := range shops {
		next <- i
	}
	close(next)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}

	s := Summary{Cheapest: map[string]Offer{}}
	for i, shop := range shops {
		if errs[i] != nil {
			s.Failed = append(s.Failed, shop.Name)
			continue
		}
		for _, e := range results[i] {
			if e.Item == "" || e.Price < 1 {
				continue
			}
			best, ok := s.Cheapest[e.Item]
			if !ok || e.Price < best.Price || (e.Price == best.Price && shop.Name < best.Shop) {
				s.Cheapest[e.Item] = Offer{Shop: shop.Name, Price: e.Price}
			}
		}
	}
	slices.Sort(s.Failed)
	return s, nil
}

// fetch reads one shop, retrying after a 5xx answer with a doubling backoff.
func fetch(ctx context.Context, client *http.Client, url string, cfg Config) ([]entry, error) {
	wait := cfg.Backoff
	for attempt := 0; ; attempt++ {
		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx); err != nil {
				return nil, err
			}
		}
		entries, err := fetchOnce(ctx, client, url, cfg.Timeout)
		if !errors.Is(err, errServer) || attempt >= cfg.Retries {
			return entries, err
		}
		if err := cfg.Sleep(ctx, wait); err != nil {
			return nil, err
		}
		wait *= 2
	}
}

// fetchOnce makes one request, with its own timeout.
func fetchOnce(ctx context.Context, client *http.Client, url string, timeout time.Duration) ([]entry, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(url, "/")+"/prices", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 500:
		return nil, fmt.Errorf("%s: %w (%d)", url, errServer, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%s: unexpected status %d", url, resp.StatusCode)
	}
	var entries []entry
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&entries); err != nil {
		return nil, fmt.Errorf("%s: bad JSON: %w", url, err)
	}
	return entries, nil
}

// sleep waits for d, or until ctx is done (Day 44).
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Report formats s for the terminal: a heading, then one line per item in alphabetical
// order, then the failed shops if there are any.
//
//	Cheapest prices:
//	sukari 1kg: Ksh180 at Duka la Kona
//	unga 2kg: Ksh230 at Mama Njeri Shop
//	Failed: Gikomba Wholesale
func Report(s Summary) string {
	var b strings.Builder
	b.WriteString("Cheapest prices:\n")
	items := make([]string, 0, len(s.Cheapest))
	for item := range s.Cheapest {
		items = append(items, item)
	}
	slices.Sort(items)
	for _, item := range items {
		o := s.Cheapest[item]
		fmt.Fprintf(&b, "%s: Ksh%d at %s\n", item, o.Price, o.Shop)
	}
	if len(s.Failed) > 0 {
		fmt.Fprintf(&b, "Failed: %s\n", strings.Join(s.Failed, ", "))
	}
	return b.String()
}
