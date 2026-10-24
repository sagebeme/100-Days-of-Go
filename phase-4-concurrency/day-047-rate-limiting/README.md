# Day 47: Rate Limiting

## The brief

Your bulk-SMS service sends chama reminders through a gateway that allows 2 messages a second per account. Go faster and it answers `429 Too Many Requests`; keep going and it blocks the account for the day. Meanwhile your own API, which shops use to queue their messages, needs the same protection from a shop whose script has a bug and calls you 500 times a second.

Three tools, from simplest to the one you'd use at work:

```
Throttle(ctx, numbers, ticker.C, send)      one send per tick: a ticker every 500ms → 2 a second

b := NewTokenBucket(2, time.Second, t0)     bursts of 2, then 1 a second
b.Allow(t0)                    →  true
b.Allow(t0)                    →  true
b.Allow(t0 + 500ms)            →  false     the next token comes at t0+1s
b.Allow(t0 + 1s)               →  true

l := NewClientLimiter(rate.Every(time.Second), 2)    the same, per shop, with golang.org/x/time/rate
l.Allow("shop-a", t0) ×3       →  true, true, false
l.Allow("shop-b", t0)          →  true      shop-b has its own bucket
```

## What you'll use

- **`time.NewTicker(d)`** sends the time on its channel `C` every `d`. Call `Stop` when you're done (`defer ticker.Stop()`) so it stops ticking.
- **Taking the clock as a parameter.** `Throttle` takes a `<-chan time.Time` rather than making its own ticker, and `Allow` takes `now`. Real code passes `ticker.C` and `time.Now()`; tests pass a channel they control and fixed times. No test waits for a real second.
- **A token bucket** holds up to `capacity` tokens and gains one every `every`. Each event takes a token; no token, no event. A full bucket allows a short burst, and the refill rate sets the long-run speed.
- **`golang.org/x/time/rate`** is the Go team's token bucket. `rate.NewLimiter(rate.Every(time.Second), 2)` is "1 a second, bursts of 2". `lim.Allow()` checks now and doesn't wait; `lim.AllowN(t, 1)` checks at time `t`; `lim.Wait(ctx)` blocks until a token is free or `ctx` is done. It's already in this repo's `go.mod`.
- **`time.Duration` arithmetic**: `now.Sub(last)` is a `Duration`, and dividing two durations gives a plain count: `now.Sub(last) / every`.

```go
lim := rate.NewLimiter(rate.Every(500*time.Millisecond), 1) // 2 a second, no bursts
for _, n := range numbers {
	if err := lim.Wait(ctx); err != nil { // waits for a token, or ctx
		return err
	}
	send(n)
}
```

## Why it matters at work

Every API you call has limits, and every API you run needs them. Going over someone else's limit gets you `429` responses and, with payment and SMS providers, a blocked account. Running your own API without limits means one buggy client can take it down for everyone, and one attacker can try thousands of passwords a minute. [reference/owasp-checklist.md](../../reference/owasp-checklist.md) asks you to rate-limit logins per IP and per account and answer `429` with `Retry-After`. In review, expect "what happens if this loop runs 10 000 times a second?" Day 61 turns today's `ClientLimiter` into a per-IP limiter in front of an HTTP API.

## Steps

1. Open `starter/ratelimiting.go`. Write `Throttle`: for each item, `select` on `<-tick` (then `send(item)`) and `<-ctx.Done()` (return `ctx.Err()`). Return `nil` after the loop.
2. Give `TokenBucket` fields: `mu sync.Mutex`, `capacity`, `tokens`, `every` and `last time.Time`. `NewTokenBucket` fills it: `tokens` starts at `capacity` and `last` at `now`.
3. In `Allow`, lock the mutex, then refill: `steps := int(now.Sub(b.last) / b.every)`. If `steps > 0`, add them to `tokens` and move `last` on by `steps` whole steps with `b.last.Add(time.Duration(steps) * b.every)`. That keeps the part of a step that hasn't finished yet.
4. Still in `Allow`: if `tokens >= capacity`, set `tokens = capacity` and `last = now`. A full bucket can't save up; its next token comes one step after now.
5. Finish `Allow`: no tokens returns false; otherwise take one and return true.
6. Give `ClientLimiter` a mutex, the rate, the burst and `clients map[string]*rate.Limiter`. In `Allow`, under the lock, find the client's limiter or make one with `rate.NewLimiter(l.r, l.burst)`. Then return `lim.AllowN(now, 1)`.
7. Run the tests from the repo root, with `-race` too, then watch a real ticker at work:

```bash
go test -race ./phase-4-concurrency/day-047-rate-limiting/starter
go run ./phase-4-concurrency/day-047-rate-limiting/starter/cmd/sms
```

## The kata: Gateway Bucket

Run a token bucket over a list of message arrival times and count what gets through. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 47
```

## When you're stuck

- **`1 messages went out before the first tick`**: you sent first and waited after. Wait for the tick, then send.
- **`Throttle is still waiting for a tick after its context was cancelled`**: add the `case <-ctx.Done():` to the same `select` as the tick.
- **`call 5: Allow(t0+1000ms) = false, want true`**: the refill isn't happening. Check you divide `now.Sub(b.last)` by `b.every`, and that you store the new `last`.
- **`Allow(t0+1999ms) = true, want false`** in "part of a step carries over": you set `last = now` on every refill, which throws away the half step. Move `last` on by whole steps only, except when the bucket is full.
- **`Allow(t0+60000ms)` lets 60 through**: cap `tokens` at `capacity`.
- **`11 of 100 calls were allowed`**, or a race report: lock the mutex for the whole of `Allow`, refill and take together.
- **`assignment to entry in nil map`**: make `clients` in `NewClientLimiter`.
- **Still stuck?** Open `solution/ratelimiting.go`, read it once, close it, then write your own version from memory.
