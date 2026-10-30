# Day 44: Context

## The brief

A matatu booking app on a phone with one bar of signal. The customer taps "search" and the app asks the fare service for a price; if that takes more than 5 seconds, it should give up and say so. Then it asks every SACCO on the route for a free seat at once. The first "yes" wins, and the other requests should stop instead of finishing work nobody will read. And if the customer closes the app, everything should stop.

That's what `context.Context` is for: one value, passed down to every slow call, that says "stop now" or "stop at 18:05:30".

```
Sleep(ctx, 5*time.Second)                         →  nil                 after 5s
Sleep(ctx, 5*time.Second), ctx cancelled at 2s    →  context.Canceled    after 2s

FetchFare(ctx, "Kencom-Rongai", 5*time.Second, fetch)   fetch answers in 1s
  →  100, nil
FetchFare(ctx, "Kencom-Rongai", 5*time.Second, fetch)   fetch takes 10s
  →  0, "fare for Kencom-Rongai: context deadline exceeded"     after 5s

FirstAvailable(ctx, []string{"Super Metro", "Embassava", "Rembo"}, check)
  →  "Super Metro", nil      the others' checks are cancelled
```

*The routes and SACCO names are illustrative, not real timetables.*

## What you'll use

- **`ctx context.Context`** is the first parameter of any function that might be slow, by convention. Start from `context.Background()` at the top of a program or test.
- **`ctx.Done()`** is a channel that's closed when the context is cancelled or its deadline passes. Put it in a `select` next to the real work. **`ctx.Err()`** then says why: `context.Canceled` or `context.DeadlineExceeded`.
- **`context.WithCancel(ctx)`** and **`context.WithTimeout(ctx, d)`** make a child context and a `cancel` func. Cancelling a parent cancels every child, never the other way round.
- **Always `defer cancel()`.** It frees the child's timer and resources even when the work finishes early. `go vet` warns when you forget.
- **`%w`** in `fmt.Errorf` wraps an error, so `errors.Is(err, context.DeadlineExceeded)` still works after you add the route name ([Day 17](../../phase-2-modelling/day-017-failed-transactions/)).

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()

select {
case r := <-result:
	return r, nil
case <-ctx.Done():
	return 0, ctx.Err() // context.DeadlineExceeded after 2s
}
```

Like yesterday, the tests run inside `synctest.Test`, so a 5-second deadline takes no real time and the tests check exactly when each function returned.

## Why it matters at work

Every request your server handles has a context (Phase 5: `r.Context()`), and it's cancelled when the client hangs up. Code that passes it down to the database and to outgoing HTTP calls stops wasted work the moment nobody is waiting. Code that ignores it keeps a slow query running for a user who left a minute ago, and under load those pile up until the server falls over. In review, a slow call without a `ctx` parameter, or a `context.Background()` deep inside a handler, gets flagged.

## Steps

1. Open `starter/contextcancel.go`. Write `Sleep`: make a timer with `time.NewTimer(d)` and `defer timer.Stop()`, then `select` on `timer.C` (return `nil`) and `ctx.Done()` (return `ctx.Err()`).
2. Write `FetchFare`: `ctx, cancel := context.WithTimeout(ctx, timeout)`, `defer cancel()`, call `fetch(ctx, route)`. If it fails, return `0, fmt.Errorf("fare for %s: %w", route, err)`.
3. In `FirstAvailable`, make `checkCtx, cancel := context.WithCancel(ctx)` and `defer cancel()`.
4. Make a buffered channel `answers := make(chan string, len(saccos))`. Start one goroutine per SACCO with a `sync.WaitGroup`. Each calls `check(checkCtx, s)` and sends `s` on `answers` if the answer is yes with no error, or `""` otherwise.
5. Receive exactly `len(saccos)` answers. The first non-empty one is the winner: remember it and call `cancel()` straight away so the other checks stop.
6. Call `wg.Wait()`, then decide: a winner returns it; otherwise if `ctx.Err()` (the parent's) isn't nil, return that; otherwise return `ErrNoneAvailable`.
7. Run the tests from the repo root:

```bash
go test ./phase-4-concurrency/day-044-context-cancel/starter
```

## The kata: Cancel Tree

Contexts form a tree, and cancelling one cancels its whole subtree. Count how many each cancel newly stops. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 44
```

## When you're stuck

- **`fetch's context is still live after FetchFare returned`**: you dropped the `cancel` func with `ctx, _ :=`. `go vet` says **`the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak`**. Add `defer cancel()`.
- **`fetch's context has deadline 0001-01-01 ...`**: you passed the original `ctx` to `fetch` instead of the new one. Use `ctx, cancel := ...` so the name now means the child.
- **`want an error wrapping context.Canceled`**: you built the timeout on `context.Background()` instead of `ctx`, so cancelling the parent never reached `fetch`.
- **`error message = ... want "fare for Kencom-Rongai: ..."`**: use `%w`, not `%v`, and match the text exactly.
- **`N checks had their context cancelled ... want M`**: either you didn't call `cancel()` when the winner arrived, or you returned before `wg.Wait()`, while the other checks were still stopping.
- **`FirstAvailable returned after 24h0m0s`**: you waited for every check to finish naturally. One that hangs only stops when its context is cancelled, so cancel as soon as you have a winner.
- **Still stuck?** Open `solution/contextcancel.go`, read it once, close it, then write your own version from memory.
