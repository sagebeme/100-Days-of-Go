# Day 42: Channels

## The brief

Baraka runs a small shop with an M-Pesa till. All day the till texts him one line per payment, mixed in with promos and the odd scam message. At closing time he wants the day's total, and his accountant wants the payments in bundles of ten for her spreadsheet.

Build it as a pipeline. One stage hands out the raw lines, the next turns good lines into payments and drops the rest, and the last stage adds them up or bundles them. Each stage runs in its own goroutine and passes values to the next over a channel.

```
Lines([]string{"QFT1ABC2DE 500", "RKD4XY7Z1P 1250"})
  →  a channel delivering "QFT1ABC2DE 500", "RKD4XY7Z1P 1250", then closed

Parse(lines)       "QFT1ABC2DE 500"       →  Payment{Code: "QFT1ABC2DE", Amount: 500}
                   "you have won a car!"  →  skipped

Total(Parse(Lines(msgs)))  →  1770

Batch(payments A B C D E, 2)  →  [A B], [C D], [E], then closed
```

*Real till messages are longer, like the one you built on Day 1. Today's are cut down to a code and an amount so the day is about channels, not parsing.*

## What you'll use

- **`make(chan T)`** makes a channel of `T`. `ch <- v` sends, `v := <-ch` receives. On an unbuffered channel a send waits until someone receives, so the two goroutines meet.
- **`make(chan T, 10)`** makes a buffered channel: sends only wait once 10 values are waiting to be received.
- **`close(ch)`** says "nothing more is coming". Receivers still get the values already sent, and then `v, ok := <-ch` gives `ok == false`.
- **`for v := range ch`** receives until the channel is closed and empty, then the loop ends.
- **Direction in the type.** `<-chan T` can only be received from, `chan<- T` can only be sent to. A function that returns `<-chan T` promises callers they only read.
- **The rule for closing:** the goroutine that sends closes, once, when it's finished. Receivers never close. Sending on a closed channel panics.

```go
func Count(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // runs when the goroutine's function returns
		for i := range n {
			out <- i
		}
	}()
	return out // returned straight away; the goroutine keeps sending
}

for x := range Count(3) {
	fmt.Println(x) // 0, 1, 2, then the loop ends because out was closed
}
```

## Steps

1. Open `starter/channels.go`. In `Lines`, make an unbuffered `chan string` called `out`.
2. Start a goroutine that sends each message on `out` and then closes it. Use `defer close(out)` as its first line. Return `out` after the `go` statement, not inside the goroutine.
3. Write `Parse` the same shape: make `out`, start a goroutine with `defer close(out)`, and `range` over `in`. Split each line with `strings.Fields`, skip it unless there are exactly two fields, convert the second with `strconv.Atoi`, and skip it if that fails or the amount is below 1. Send the `Payment` that's left.
4. Write `Total`. It returns a number, not a channel, so it needs no goroutine: `range` over `in` and add up.
5. Write `Batch`. Collect payments into a slice; when it reaches `size`, send it and start a new slice with `batch = nil`. When `in` is closed, send what's left if it isn't empty. Start with `size = max(size, 1)`.
6. Run the tests from the repo root:

```bash
go test ./phase-4-concurrency/day-042-channels/starter
```

## The kata: Round Robin

An M-Pesa agent serves a queue of customers in turns of at most `q` minutes, sending unfinished customers to the back. Print who finishes when. It's the classic round-robin scheduling problem, and a queue is exactly what a channel is. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 42
```

## When you're stuck

- **`Lines didn't return`**: you send on `out` before returning it. Nobody can receive until the function returns, so the first send waits forever. Do the sending inside `go func() { ... }()`.
- **`fatal error: all goroutines are asleep - deadlock!`**: the same mistake outside a test, or a `range` over a channel nobody closes. Go found that every goroutine is waiting on another.
- **`...'s channel was never closed`**: add `defer close(out)` at the top of the sending goroutine. Without it, the receiver's `range` loop never ends.
- **`panic: send on closed channel`**: you closed `out` before the goroutine finished sending, often by writing `close(out)` after the `go` statement instead of inside the goroutine.
- **`panic: close of closed channel`**: two places close the same channel. Only the one sender closes, once.
- **Earlier batches change into later ones**: you reset with `batch = batch[:0]`. That reuses the array the receiver already holds, so the next appends overwrite its payments. Start a fresh slice with `batch = nil` after each send.
- **`Batch` sends an empty `[]` at the end**: only send the leftover batch if `len(batch) > 0`.
- **Still stuck?** Open `solution/channels.go`, read it once, close it, then write your own version from memory.
