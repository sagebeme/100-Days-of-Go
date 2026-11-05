# Phase 4 recap: Concurrency

Days 41 to 48. You learnt to do many things at once without losing track of them: start work, wait for it, cancel it, share state safely, and stay polite to the servers you call.

## What the phase covered

| Days | Topic |
|---|---|
| 41 | goroutines and `sync.WaitGroup` |
| 42 | channels: sending, receiving, closing and ranging |
| 43 | `select`: timeouts, several channels at once, `default` |
| 44 | `context`: cancellation and deadlines |
| 45 | shared state, `sync.Mutex` and the race detector |
| 46 | worker pools: bounded concurrency, fan-out and fan-in |
| 47 | tickers and token buckets |
| 48 | the capstone: Price Watcher, scraping many shops at once without hammering any |

## Key ideas

1. **Every goroutine you start needs a way to finish.** Know who waits for it (`WaitGroup`), and how it's told to stop (`ctx.Done()` or a closed channel). A goroutine stuck for ever is a leak.
2. **The sender closes the channel.** Closing tells receivers "no more"; `range` stops, and a receive gives the zero value and `false`. Sending on a closed channel panics.
3. **`select` waits on several things at once.** Add a `case <-ctx.Done()` to anything that can block, and `default` only when you mean "don't wait".
4. **Pass `ctx` as the first argument and respect it.** Cancellation flows down the call tree; a deadline set at the top reaches every HTTP call and query below.
5. **Share memory by communicating, or lock it.** Either one goroutine owns the data and others send it messages, or every access holds the same mutex. `go test -race` finds the places you forgot.
6. **Bound your concurrency.** A worker pool of N goroutines reading from one jobs channel keeps memory, sockets and the other side's patience under control.
7. **Rate limits protect the services you call.** A token bucket allows bursts but caps the average; back off and retry on failure instead of hammering.
8. **Collect results without racing.** Give each worker its own slot in a slice, or send results down a channel, then merge after `Wait`.

## Self-test

1. Predict the output:

   ```go
   ch := make(chan int, 2)
   ch <- 1
   ch <- 2
   close(ch)
   for v := range ch {
       fmt.Print(v, " ")
   }
   v, ok := <-ch
   fmt.Println(v, ok)
   ```

2. What does this print, and after how long?

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
   defer cancel()
   select {
   case <-time.After(time.Second):
       fmt.Println("done")
   case <-ctx.Done():
       fmt.Println(ctx.Err())
   }
   ```

3. Ten goroutines each do `count++` on a shared `int` 1000 times. Why isn't the total always 10 000, and name two fixes.
4. A program starts one goroutine per URL for 50 000 URLs. What goes wrong, and what shape fixes it?
5. Why call `wg.Add(1)` before the `go` statement rather than inside the goroutine?

<details>
<summary>Answers</summary>

1. `1 2 0 false`: `range` drains the buffer then stops at the close, and a receive on a closed, empty channel gives the zero value and `false`.
2. `context deadline exceeded`, after about 20 ms.
3. `count++` is a read, an add and a write, and two goroutines can interleave them and lose updates: a data race. Fix it with a `sync.Mutex` around the increment, or `atomic.Int64`, or by giving each goroutine its own counter and adding them up after `Wait`.
4. Thousands of open connections at once: you run out of file descriptors or memory, and the sites see a flood. Use a worker pool: a fixed number of workers reading URLs from a channel, with a rate limiter per site.
5. If the goroutine calls `Add` itself, `Wait` can run first, see a count of zero and return before the work has started.

</details>

## Build it without steps: Till Uptime Checker

A SACCO has 40 branches, and each branch's till server answers `GET /health`. The IT officer wants a command that checks them all at once every morning and says which are down, without taking all morning and without knocking over the slow ones.

Use only Phases 1 to 4: the standard library and `golang.org/x/time/rate`.

Done when:

- `uptime -workers 8 -timeout 2s branches.txt` reads one URL per line and checks them with at most 8 requests in flight.
- Each request has its own timeout through `context`, and a single slow branch never delays the report past the timeout plus a little.
- No host gets more than 2 requests a second, even when its URL is listed twice, and a failed check is retried twice with a doubling backoff.
- Ctrl+C cancels everything still running, prints what's known so far, and exits with status 130.
- The report lists down branches first, then slow ones (over 1 s), then healthy ones, each sorted by name, so it's the same on every run. Exit status is 1 if any branch is down.
- Tests run against `httptest` servers that are fast, slow, failing or flaky, and pass with `go test -race`.
- No goroutine is left running when the command returns: a test checks it with `runtime.NumGoroutine` before and after.
