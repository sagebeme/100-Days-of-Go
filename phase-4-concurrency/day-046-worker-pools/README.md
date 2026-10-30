# Day 46: Worker Pools

## The brief

At the end of every month the Umoja chama sends each of its 2 000 members a PDF statement. Rendering one takes a second and a lot of memory. Start 2 000 goroutines at once ([Day 41](../day-041-goroutines/)) and the small server runs out of memory; render them one by one and it takes half an hour. The answer is in between: a fixed pool of workers, say 4, that take jobs from a queue until it's empty.

Today you write three tools you'll use for the rest of the course:

```
Process([]string{"Akinyi", "Mutua", "Chebet"}, 2, render)
  →  []string{"statement for Akinyi", "statement for Mutua", "statement for Chebet"}
     never more than 2 renders at once, results in job order

FanIn(smsResults, emailResults, whatsappResults)
  →  one channel with every value from all three, closed when they all are

g, ctx := WithContext(ctx)
g.SetLimit(4)
for _, m := range members { g.Go(func() error { return send(ctx, m) }) }
err := g.Wait()   →  the first error, and ctx was cancelled so the other sends could stop
```

## What you'll use

- **A worker pool**: start `k` goroutines that each `range` over one jobs channel. Send the jobs, then `close` the channel; each worker's loop ends when no jobs are left. `k` is the most that ever run at once.
- **Fan-out, fan-in**: fan-out is many goroutines reading one channel. Fan-in is the reverse: one goroutine per input forwards into a single output, and a last goroutine closes the output after a `WaitGroup` says they're all done.
- **A semaphore channel**: `slots := make(chan struct{}, n)`. Sending takes a slot and waits when all `n` are taken; receiving gives one back. `struct{}` takes no memory.
- **`sync.Once`**: `once.Do(f)` runs `f` the first time only, however many goroutines call it. Here it keeps the first error.
- **Generics** ([Day 18](../../phase-2-modelling/day-018-typed-inventory/)): `Process[J, R any]` works for any job and result type.

```go
jobs := make(chan int)
var wg sync.WaitGroup
for range 4 { // 4 workers
	wg.Go(func() {
		for j := range jobs {
			render(j)
		}
	})
}
for j := range 2000 {
	jobs <- j
}
close(jobs)
wg.Wait()
```

The `Group` you build is a small copy of `errgroup.Group` from `golang.org/x/sync/errgroup`, which is what most Go teams use. After today it holds no mystery.

## Steps

1. Open `starter/workerpools.go`. In `Process`, make `results := make([]R, len(jobs))` and an unbuffered `next := make(chan int)` for job numbers.
2. Start `max(workers, 1)` workers with a `sync.WaitGroup`. Each does `for i := range next { results[i] = work(jobs[i]) }`.
3. Send every index on `next`, `close(next)`, `wg.Wait()`, return `results`.
4. Write `FanIn`: make `out`. For each input start a goroutine (counted by a `WaitGroup`) that forwards with `for v := range ch { out <- v }`. Start one more goroutine that does `wg.Wait()` and then `close(out)`. Return `out`.
5. Give `Group` these fields: `wg sync.WaitGroup`, `slots chan struct{}`, `once sync.Once`, `err error`, `cancel context.CancelFunc`.
6. `WithContext` makes a child with `context.WithCancel` and returns `&Group{cancel: cancel}` and the child. `SetLimit(n)` makes `slots` with room for `n`, or sets it to `nil` when `n < 1`.
7. `Go`: if `slots` isn't nil, send on it first (this is where Go waits). Then start the task with `g.wg.Go`. Inside, `defer` taking a value back off `slots`, run `f`, and on an error use `g.once.Do` to save it and call `g.cancel` if it isn't nil.
8. `Wait`: `g.wg.Wait()`, call `g.cancel` if it isn't nil, return `g.err`.
9. Run the tests from the repo root, with the race detector too:

```bash
go test ./phase-4-concurrency/day-046-worker-pools/starter
go test -race ./phase-4-concurrency/day-046-worker-pools/starter
```

## The kata: Statement Workers

`k` workers take `n` jobs in order, each job going to the first worker to come free. When is everything done? It's the classic "parallel processing" problem, and your first use of a heap. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 46
```

## When you're stuck

- **`N jobs were running at once, want 3`** with N bigger than 3: you started a goroutine per job. Start exactly `workers` goroutines and let them share the jobs channel.
- **`1 jobs were running at once, want 3`**: the workers share a lock around `work`, or you process jobs in the sending loop. Only the hand-out goes through the channel; `work` runs in each worker freely.
- **`fatal error: all goroutines are asleep - deadlock!`** in `Process`: you sent the jobs before starting the workers, on an unbuffered channel. Start the workers first.
- **The test hangs, or `panic: deadlock: main bubble goroutine has exited but blocked goroutines remain`**: you forgot `close(next)`, so the workers wait for jobs forever and `wg.Wait()` never returns.
- **`panic: send on closed channel`** in `FanIn`: you closed `out` before the forwarders finished. Close it in its own goroutine, after `wg.Wait()`.
- **`Go had returned 6 times with 2 slots full`**: take the slot in `Go` itself, before starting the goroutine, not inside it.
- **`the group's context is still live after Wait returned`**: call `g.cancel()` in `Wait` too.
- **Still stuck?** Open `solution/workerpools.go`, read it once, close it, then write your own version from memory.
