# Day 45: Mutexes

## The brief

The Umoja chama has 2 000 members, and on contribution day the M-Pesa callbacks arrive all at once. Your server handles each one in its own goroutine. Last month's total came out Ksh150 short and nobody could say why; run it again and it's a different number.

The culprit is already in `starter/mutexes.go`: `TotalContributions` has a data race. Today you catch it with Go's race detector and fix it with a `sync.Mutex`. Then you build the chama's `Ledger`, which must stay right with any number of goroutines depositing and withdrawing at once.

```
TotalContributions([]int{500, 1000, 250})   →  1750, every time

l := NewLedger()
l.Deposit("Akinyi", 500)
l.Withdraw("Akinyi", 300)     →  nil                     Balance("Akinyi") is 200
l.Withdraw("Akinyi", 300)     →  ErrInsufficientFunds    still 200

50 goroutines each Withdraw("chama", 100) from Ksh1000   →  exactly 10 succeed, balance 0
```

## What you'll use

- **A data race** is two goroutines touching the same variable at the same time, with at least one writing. `total += a` is really load, add, store. Two goroutines can both load 1000, both add, and both store their own answer: one payment vanishes.
- **`sync.Mutex`** lets one goroutine at a time through. `mu.Lock()` waits until the mutex is free and takes it; `mu.Unlock()` gives it back. Code between them is a critical section.
- **`defer mu.Unlock()`** right after `Lock` unlocks however the function returns, including early returns.
- **A mutex in a struct** guards the fields next to it. Don't copy the struct after first use, so use pointer receivers ([Day 15](../../phase-2-modelling/day-015-boda-fleet/)). `go vet` warns about copies: `passes lock by value`.
- **The race detector.** `go test -race` builds your code with extra checks that watch every memory access and report a race when they see one. It slows code down a lot, so it's for tests, not production.

```go
type Tally struct {
	mu sync.Mutex
	n  map[string]int
}

func (t *Tally) Add(stage string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.n[stage]++
}
```

When reads far outnumber writes, `sync.RWMutex` lets many readers in at once with `RLock`/`RUnlock`, while `Lock` still waits for everyone. For a single number, `sync/atomic` (`atomic.Int64` and its `Add` method) is a lighter fix. Start with a plain `Mutex`; it's the easiest to get right.

## Why it matters at work

A data race doesn't fail every time. It passes on your laptop, passes in CI, and loses money in production on the busiest day of the month, when goroutines really do run at once. That's why serious Go teams run `go test -race` in CI ([Day 63](../../phase-5-backend/day-063-docker-and-ci/) sets that up), and why a reviewer who sees a map or counter shared between goroutines asks "what guards this?". The answer should be a mutex in the same struct, or a channel that owns the data.

## Steps

1. Run the tests with the race detector from the repo root:

```bash
go test -race ./phase-4-concurrency/day-045-mutexes/starter
```

2. Read the first report. It looks like this:

```
WARNING: DATA RACE
Read at 0x00c00011c1f8 by goroutine 11:
  .../day-045-mutexes/starter.TotalContributions.func1()
      .../day-045-mutexes/starter/mutexes.go:22 +0x33

Previous write at 0x00c00011c1f8 by goroutine 10:
  .../day-045-mutexes/starter.TotalContributions.func1()
      .../day-045-mutexes/starter/mutexes.go:22 +0x4a
...
--- FAIL: TestTotalContributions/three_members (0.00s)
    testing.go:1867: race detected during execution of test
```

   Two goroutines, the same address, line 22: `total += a`. Run the test without `-race` a few times too. Sometimes it passes, sometimes the total is short. That's why you need the detector.
3. Fix `TotalContributions`: declare `var mu sync.Mutex` next to `total`, and wrap `total += a` in `mu.Lock()` and `mu.Unlock()`. Keep the goroutines.
4. Give `Ledger` two fields: `mu sync.Mutex` and `balances map[string]int`. Make the map in `NewLedger`.
5. Write `Deposit`, `Balance` and `Total`. Each starts with `l.mu.Lock()` and `defer l.mu.Unlock()`. Reading needs the lock too: a read during a write is still a race.
6. Write `Withdraw`. The check (`balance < amount`) and the change must be inside the same lock. If you unlock between them, two goroutines can both pass the check before either takes the money.
7. Run the tests again, with and without `-race`, until both pass:

```bash
go test ./phase-4-concurrency/day-045-mutexes/starter
go test -race ./phase-4-concurrency/day-045-mutexes/starter
```

## The kata: Lost Updates

Simulate `x++` from several goroutines, step by step, following a given schedule, and count the increments that vanish. It's today's race with the scheduler written down. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 45
```

## When you're stuck

- **`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`**: the race detector needs a C compiler. Install `gcc` (on Windows, through MSYS2 or WSL) and run with `CGO_ENABLED=1`.
- **`fatal error: concurrent map writes`** or **`concurrent map read and map write`**: Go crashes on purpose when it catches goroutines using a map without a lock. Every method that touches `balances` must hold `l.mu`.
- **`assignment to entry in nil map`**: make the map in `NewLedger` with `map[string]int{}`.
- **`11 withdrawals succeeded`** or a negative balance: you checked the balance and changed it under two separate locks. Hold one lock for both.
- **`passes lock by value`** from `go vet`: a method has a value receiver `(l Ledger)`, which copies the mutex. Use `(l *Ledger)`.
- **The test hangs**: you called `Lock` twice without `Unlock` in between, for example `Total` calling `Balance` while holding the lock. A Go mutex can't be locked again by the same goroutine. Read the map directly instead.
- **Still stuck?** Open `solution/mutexes.go`, read it once, close it, then write your own version from memory.
