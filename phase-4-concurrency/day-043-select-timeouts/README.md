# Day 43: Select

## The brief

Your airtime app asks two networks for the price of a 1 GB bundle at the same time and shows whichever answers first. If neither answers in 5 seconds, it says so instead of spinning forever.

The same app books matatu seats. When the matatu is full, the customer should hear "full" at once, not wait for a seat to free up. And the booking office has two ticket counters whose sales need to land in one queue.

```
Fastest(saf, air, 5*time.Second)   air answers after 2s, saf never  →  Quote{"Airtel", 95}, nil     after 2s
Fastest(saf, air, 5*time.Second)   neither answers                  →  Quote{}, ErrTimeout         after 5s

TryBook(seats, "Achieng")          a seat is free                   →  true
TryBook(seats, "Njeri")            the matatu is full               →  false, straight away

Merge(counterA, counterB)          →  every ticket from both, then closed once both are closed
```

## What you'll use

- **`select`** waits on several channel operations and runs the first one that can go ahead. If several are ready, it picks one at random.
- **A timeout** is just one more case. `time.After(d)` returns a channel that receives once after `d`. `time.NewTimer(d)` does the same and lets you `Stop` it when you finish early.
- **`default`** runs when no other case is ready, so the `select` never waits. That's a non-blocking send or receive.
- **The nil-channel trick.** A receive from a closed channel never waits: it gives the zero value with `ok == false`, every time. A receive from a `nil` channel waits forever. So once a channel is closed, set your variable to `nil` and `select` stops picking that case.

```go
select {
case q := <-quotes:
	fmt.Println("got", q)
case <-time.After(5 * time.Second):
	fmt.Println("gave up")
}

select {
case seats <- name: // there was room
default: // full: carry on without waiting
}
```

The tests for `Fastest` run inside `synctest.Test` from the standard library's `testing/synctest`. In there, time is fake: it jumps ahead whenever every goroutine is waiting. A 5-second timeout takes no real time and is exactly 5 seconds long, so the test can check when `Fastest` returned without being slow or flaky.

## Steps

1. Open `starter/selecttimeouts.go`. In `Fastest`, make a timer with `timer := time.NewTimer(timeout)` and `defer timer.Stop()`.
2. Loop `for a != nil || b != nil`. Inside, `select` on three cases: receive `q, ok := <-a`, receive `q, ok := <-b`, and `<-timer.C`.
3. For the quote cases: if `ok` is false, set that channel to `nil` and `continue`. Otherwise return the quote. The timer case returns `ErrTimeout`.
4. After the loop (both channels closed), return `ErrNoQuotes`.
5. Write `TryBook` as one `select`: send `name` on `seats` and return true, or `default` and return false.
6. Write `Merge`: make `out`, start a goroutine with `defer close(out)`, and loop like `Fastest` does, but send each value on `out` instead of returning it. Return `out`.
7. Run the tests from the repo root:

```bash
go test ./phase-4-concurrency/day-043-select-timeouts/starter
```

## The kata: Impatient Queue

Customers at a SACCO counter each wait for the teller or for their own deadline, whichever comes first. Count who gets served. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 43
```

## When you're stuck

- **`Fastest(5s timeout) was still waiting an hour later`**: there's no timeout case in your `select`, or you made a new `time.After` channel each time round the loop, which restarts the clock. Make the timer once, before the loop.
- **`Fastest returned after 0s, want 4s`** with an empty quote: you returned the zero value from a closed channel. Check `ok`, and set the channel to `nil` when it's false.
- **`panic: deadlock: main bubble goroutine has exited but blocked goroutines remain`**: something you started is still waiting when the test ends, often a goroutine sending on a channel nobody reads. The stack trace under the panic shows where it's stuck.
- **`TryBook ... waited instead of returning false`**: add the `default` case.
- **`Merge sent more values than went in: ["A1" "" "" ""...]`**: a closed input keeps giving `""`. Set it to `nil` when `ok` is false.
- **`Merge's channel was never closed`**: the loop must stop when both inputs are `nil`, and `defer close(out)` must be in the goroutine.
- **Still stuck?** Open `solution/selecttimeouts.go`, read it once, close it, then write your own version from memory.
