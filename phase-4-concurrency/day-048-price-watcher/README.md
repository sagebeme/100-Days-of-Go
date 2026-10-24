# Day 48: Price Watcher

## The brief

Mama Njeri runs a small shop in Kawangware and restocks every week. She wants to know, for every item, which wholesaler near her is cheapest today. Each wholesaler has a little price service: `GET /prices` answers with a JSON list. Some are quick. Some are on a slow connection. One has a server that falls over for a second now and then, and one returns rubbish.

Build the watcher. It reads every shop at once, but never more than a few requests at a time, never faster than a global rate limit, never waiting forever on one shop, and it tries again when a shop's server has a bad moment.

This is the Phase 4 capstone. There is no walkthrough today: you get the brief, the requirements, the tests and some hints. Everything you need is from Days 41 to 47, plus a few lines of HTTP client code shown below.

```
GET http://127.0.0.1:41235/prices
  →  200  [{"item":"unga 2kg","price":230},{"item":"sukari 1kg","price":190}]

Watch(ctx, http.DefaultClient, shops, Config{Workers: 3, Timeout: 500*time.Millisecond, Retries: 3, Backoff: 100*time.Millisecond, Limiter: lim})
  →  Summary{
       Cheapest: {"unga 2kg": {"Kariakor Traders", 228}, "sukari 1kg": {"Ngara Stores", 175}, ...},
       Failed:   ["Gikomba Wholesale", "Zawadi Mart"],
     }, nil

Report(summary)
  →  Cheapest prices:
     chumvi 500g: Ksh30 at Duka la Kona
     mafuta 1L: Ksh310 at Kariakor Traders
     sukari 1kg: Ksh175 at Ngara Stores
     unga 2kg: Ksh228 at Kariakor Traders
     Failed: Gikomba Wholesale, Zawadi Mart
```

*The shops are made up, and so are their prices. Every shop in the tests and in `cmd/watch` is a fake server started inside the program with `httptest`; nothing goes to the internet.*

## What you'll use

Phase 5 is all about HTTP. Today you only need the client side, and only this much:

```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, shop.URL+"/prices", nil)
if err != nil {
	return err
}
resp, err := client.Do(req) // a network error or a timeout comes back here
if err != nil {
	return err
}
defer resp.Body.Close() // always, or the connection is never reused
if resp.StatusCode != http.StatusOK {
	// 5xx: the shop's fault, worth retrying. 4xx: ours, retrying won't help.
}
var entries []struct {
	Item  string `json:"item"`
	Price int    `json:"price"`
}
err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&entries) // read at most 1 MB
```

[reference/http.md](../../reference/http.md) has the full list of status codes. `httptest.NewServer(handler)` starts a real server on `127.0.0.1` with a random port and gives you its `URL`; the tests use it for every fake shop.

`Config.Limiter` is an interface with one method, `Wait(ctx) error`. A `*rate.Limiter` from Day 47 already has that method, so it fits without any glue, and the tests pass a fake that counts calls.

## Why it matters at work

Almost every backend calls other services: M-Pesa, an SMS gateway, a partner's API. Each call can be slow, fail, or fail only sometimes. The habits here are what reviewers look for: a timeout on every outgoing request, a bound on how many run at once, a rate limit you control, retries only for errors that can go away (5xx, not 4xx), with a growing wait between them so a struggling server gets room to recover, and one bad dependency never taking the whole result down with it.

## Steps

Requirements. Meet them in any order. `go test` checks each one.

1. `Watch` reads `GET <shop.URL>/prices` for every shop and returns, for each item, the cheapest `Offer`. When two shops share the lowest price, the shop whose name sorts first wins.
2. Entries with an empty item name or a price below 1 are ignored.
3. At most `cfg.Workers` requests are in flight at once (below 1 counts as 1). Use a worker pool, not a goroutine per shop.
4. Every attempt has its own timeout of `cfg.Timeout` (0 means none), made with `context.WithTimeout` from the `ctx` that `Watch` got.
5. If `cfg.Limiter` isn't nil, call its `Wait(ctx)` before every attempt, retries included. If `Wait` returns an error, that shop has failed.
6. A `5xx` answer is retried, up to `cfg.Retries` extra attempts. Before retry 1 wait `cfg.Backoff`, before retry 2 twice that, and so on, using `cfg.Sleep`. If `cfg.Sleep` is nil, wait for real with a timer and `ctx`, as on Day 44.
7. Anything else that goes wrong (a timeout, a network error, a `4xx` including `429`, bad JSON) fails that shop at once, with no retry.
8. A shop that fails goes into `Summary.Failed`, sorted by name. It never stops the other shops, and it isn't an error from `Watch`.
9. `Watch` returns an error only when `ctx` itself is done, and then it returns `ctx.Err()`.
10. `Report` prints `Cheapest prices:`, then `item: Ksh<price> at <shop>` for each item in alphabetical order, then `Failed: ` and the failed shops joined with `, `, but only if there are any. Every line ends with `\n`.
11. Run the tests from the repo root, with `-race` too, then run it against six fake shops:

```bash
go test -race ./phase-4-concurrency/day-048-price-watcher/starter
go run ./phase-4-concurrency/day-048-price-watcher/starter/cmd/watch
```

When it works, make it yours: add a `-workers` flag to `cmd/watch` (Day 22), or print how many attempts each shop needed.

## The kata: Price Board

Shops change prices and drop items all day; answer "cheapest right now?" fast. A heap with lazy deletion does it. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 48
```

## When you're stuck

- **Where to start**: split it in three. `fetchOnce` makes one request and returns the entries or an error. `fetch` wraps it in the retry loop with the limiter and backoff. `Watch` runs `fetch` from a Day 46 worker pool, storing each shop's entries and error in slot `i`, then merges once every worker is done.
- **`12 requests were open at once with Workers: 3`**: you started a goroutine per shop. Start `cfg.Workers` goroutines that `range` over a channel of shop numbers.
- **`Watch took 3s with a 100ms timeout`**: the request isn't using the timeout context. Build the request with `http.NewRequestWithContext(ctx, ...)` using the context from `context.WithTimeout`.
- **`the shop got 4 requests, want 1`** for a 404 or 429: retry only when `resp.StatusCode >= 500`. A wrapped sentinel error (`var errServer = errors.New(...)`, then `errors.Is`) is a tidy way for `fetchOnce` to tell `fetch` "this one is worth retrying".
- **`backoff waits = [10ms 10ms], want [10ms 20ms]`**: double the wait after each retry, `wait *= 2`.
- **`Limiter.Wait was called 3 times, want 4`**: call it inside the retry loop, before every attempt, not once per shop.
- **`Watch error = Get "http://127.0.0.1:1/prices": dial tcp ... connection refused, want nil`**: one shop's failure goes into `Failed`. Only `ctx.Err()` is an error from `Watch`.
- **`Failed = ["Moved Away" "Broken JSON" ...]`**: sort it with `slices.Sort`.
- **A race report on a map**: don't write the `Cheapest` map from the workers. Keep per-shop results in a slice slot each, and merge after `wg.Wait()`.
- **Still stuck?** Open `solution/pricewatcher.go`, read it once, close it, then write your own version from memory.
