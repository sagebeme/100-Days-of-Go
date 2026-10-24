# Day 41: Goroutines

## The brief

Wanjiru is treasurer of a 30-member chama. Before every meeting she sends each member a reminder through an SMS gateway, and each send takes about a second while the gateway answers. One after another, that's half a minute of waiting. Then she checks the chama's money, which sits in an M-Pesa till, a bank account and a SACCO, and each balance lookup is slow too.

None of these calls needs the others. Start them all at once, wait for them all to finish, and the whole job takes about as long as the slowest call.

```
SendAll([]string{"0712345678", "0723456789"}, send)
  →  []string{"0712345678: sent", "0723456789: sent"}     both sends run at the same time

SumBalances([]string{"mpesa", "bank", "sacco"}, balance)
  →  85500                                                 12 500 + 60 000 + 13 000
```

`send` and `balance` are passed in as functions, so the tests can stand in for the real gateway and the real banks. You met function values on Day 10 with `sort.Slice`.

## What you'll use

- **`go f()`** starts `f` in a new goroutine: a function running at the same time as the rest of your program. The `go` statement returns at once; it doesn't wait for `f`.
- **`sync.WaitGroup`** counts goroutines that haven't finished. `wg.Go(f)` starts `f` in a goroutine and counts it. `wg.Wait()` blocks until every counted goroutine has returned.
- **The older form** you'll see in most code written before Go 1.25: `wg.Add(1)` before `go`, and `defer wg.Done()` as the first line inside the goroutine.
- **One slot each.** Two goroutines must never write the same variable at once. Writing to different elements of a slice is fine, so give goroutine `i` the slot `results[i]`. Day 45 covers what happens when goroutines do share a variable.

```go
results := make([]string, len(numbers))
var wg sync.WaitGroup
for i, n := range numbers {
	wg.Go(func() {
		results[i] = send(n) // the closure sees this iteration's i and n
	})
}
wg.Wait() // every slot is filled after this line
```

Since Go 1.22 each loop iteration gets its own `i` and `n`, so the closure above is safe. Older articles copy them first (`i, n := i, n`) because every goroutine used to share one variable.

When `main` returns, the program ends and any goroutines still running are killed mid-step. That's why you always wait.

## Steps

1. Open `starter/goroutines.go` and add `import "sync"` under the `package` line.
2. In `SendAll`, make a `results` slice with `make([]string, len(numbers))`. The slots exist before any goroutine starts.
3. Declare `var wg sync.WaitGroup`. Loop over `numbers` with `for i, n := range numbers` and, inside, call `wg.Go` with a function that stores `send(n)` in `results[i]`.
4. After the loop, call `wg.Wait()`, then return `results`.
5. Write `SumBalances` the same way, with an `amounts` slice of `int`. This time use the older form, `wg.Add(1)`, `go func() { defer wg.Done(); ... }()`, so you can read both styles at work.
6. After `wg.Wait()`, add up `amounts` in a normal loop and return the total. Don't add inside the goroutines: they would all be writing one `total` at the same time.
7. Run the tests from the repo root:

```bash
go test ./phase-4-concurrency/day-041-goroutines/starter
```

The tests hold every call until all of them have started. Code that sends one at a time can't pass, however fast it is.

## The kata: Batch Sender

A gateway sends messages `k` at a time and waits for each batch's slowest message before starting the next. Work out the time saved. It's the arithmetic behind `wg.Wait()`. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 41
```

## When you're stuck

- **`SendAll returned [] ...` or the results are all `""`**: you returned before the goroutines finished. Call `wg.Wait()` before `return`.
- **`the sends never all ran at the same time`**: you called `send(n)` directly in the loop, or called `wg.Wait()` inside the loop. Start every goroutine first, then wait once, after the loop.
- **`sync: negative WaitGroup counter`**: `Done` was called more times than `Add`. With `wg.Go` you don't call `Add` or `Done` at all.
- **The test hangs, then `panic: test timed out after 10m0s`**: `Add` and `Done` don't match the other way: `Wait` is still waiting for a `Done` that never comes. Put `wg.Add(1)` right before each `go`, and `defer wg.Done()` at the top of the goroutine.
- **`fatal error: all goroutines are asleep - deadlock!`**: the same mismatch in a program with nothing else running. Go noticed that nothing can ever wake `Wait` up.
- **`undefined: sync`**: add `import "sync"`.
- **Still stuck?** Open `solution/goroutines.go`, read it once, close it, then write your own version from memory.
