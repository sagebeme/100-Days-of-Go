# Day 98: Make It 10x Faster

## The brief

Duka Books (a made-up bookkeeping app for small shops) lets an owner upload a year of M-Pesa statements and see who they paid most. A busy hardware shop in Gikomba has 60 000 lines a year, and the summary page now takes over a minute. The servers are maxed out every month-end. The developer who wrote `Summarise` has left. Musa, the CTO, says: "Don't rewrite the app. Make that one function at least 10 times faster, prove it with numbers, and don't change a single result."

```
Summarise("receipt,date,counterparty,amount\nSJK2X1ABCD,2026-10-01 08:15,MAMA MBOGA KAWANGWARE,-350.00\n...")
  →  [{KPLC PREPAID 1 0 100000} {MAMA MBOGA KAWANGWARE 2 0 47050} ...], nil

BenchmarkSummarise   before:  1 386 484 623 ns/op   181 613 242 B/op   3 155 608 allocs/op
                     after:       3 338 821 ns/op        37 480 B/op          20 allocs/op
```

**Acceptance criteria**

- `Summarise` returns exactly what it returns today, for every input, including errors and their line numbers. The correctness tests already pass; keep them passing.
- At least 10 times faster on `BenchmarkSummarise`, shown with `benchstat` over several runs.
- The allocation guards pass: at most 200 allocations for 2 000 lines, and no growth in allocations as lines are added. These count allocations rather than time, so they give the same answer on every machine.

## What you'll use

- Big-O and counting operations, from [Day 27: How Fast Is Fast](../../phase-3-algorithms/day-027-how-fast-is-fast/) and [big-o.md](../../reference/big-o.md).
- Maps for O(1) lookups, from [Day 28: Duplicate Receipts](../../phase-3-algorithms/day-028-duplicate-receipts/).
- Sorting once with a custom compare, from [Day 31](../../phase-3-algorithms/day-031-sorting-showdown/).
- New today: benchmarks and profiles. A benchmark is a test function that runs code many times and reports the cost of one run:

```go
func BenchmarkSummarise(b *testing.B) {
	in, _ := statement(5000, 200)
	b.ReportAllocs() // also report bytes and allocations per run
	for b.Loop() {   // Go 1.24: runs the body as often as needed for a steady number
		Summarise(in)
	}
}
```

`go test -bench` runs it. Add `-cpuprofile` or `-memprofile`, and `go tool pprof` shows where the time or memory goes, function by function and line by line. `testing.AllocsPerRun` counts allocations in one call, which is how the guards in the tests work.

## Steps

There's no walkthrough. Here's the method; the requirements are the acceptance criteria.

**Must-haves**

1. Run the tests and see the correctness tests pass and the two allocation guards fail. Read `Summarise` and write down every slow thing you can see before you profile. Then check your guesses.
2. Measure the starting point. Save it, because you'll compare against it:

```bash
go test -run '^$' -bench . -benchtime 3x -count 6 ./phase-8-portfolio/day-098-make-it-10x-faster/starter | tee old.txt
```

3. Profile, then read the top of the list:

```bash
cd phase-8-portfolio/day-098-make-it-10x-faster/starter
go test -run '^$' -bench . -benchtime 3x -cpuprofile cpu.out -memprofile mem.out
go tool pprof -top -cum cpu.out                   # where the CPU time goes, your functions included
go tool pprof -list 'Summarise$' cpu.out          # the time spent on each line of Summarise
go tool pprof -top -sample_index=alloc_space mem.out   # who allocates the most
go tool pprof -http=:8080 cpu.out                 # a flame graph in the browser
cd -
```

Plain `-top` sorts by the time spent inside each function itself, so the top rows are usually Go's own memory allocator. `-cum` counts the functions they call too, which brings `Summarise` and whatever it calls to the top.

4. Fix the biggest cost first, re-run the tests, and measure again. One change at a time, so you know what each change bought you. Stop when the tests pass and the benchmark is 10x faster.
5. Compare with `benchstat` (install it once with `go install golang.org/x/perf/cmd/benchstat@latest`):

```bash
go test -run '^$' -bench . -benchtime 3x -count 6 ./phase-8-portfolio/day-098-make-it-10x-faster/starter | tee new.txt
benchstat old.txt new.txt
```

   It prints the change in `sec/op`, `B/op` and `allocs/op`, with a `p=` value that says whether the difference is real or noise. Close other programs while you measure.
6. Run the tests, and keep `old.txt` and `new.txt` for the write-up. Delete `cpu.out`, `mem.out` and `starter.test` (the `.gitignore` covers `*.out` and `*.test`, but tidy anyway).

```bash
go test ./phase-8-portfolio/day-098-make-it-10x-faster/starter
```

**Stretch goals**

- Stream it: take an `io.Reader` and use a `bufio.Scanner`, so a 1 GB statement doesn't need 1 GB of memory. Benchmark both versions.
- Run `go test -bench . -benchmem` before and after a "clever" change that turns out to be slower, and keep it in your write-up. Showing a measured dead end is good engineering.
- Profile a running server instead of a benchmark: import `net/http/pprof` on a separate admin port, never the public one, and fetch `/debug/pprof/profile?seconds=10`.

## The kata: Exact Budget

Count the runs of consecutive days whose totals add up to exactly `K`. The obvious double loop is O(N²); a running sum and a map make it O(N). It's LeetCode's "subarray sum equals K", and the same idea as today: replace an inner loop with a lookup. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 98
```

## When you're stuck

- **The profile is all `regexp.compile`**: the pattern is compiled again for every line. Compile it once in a package-level `var`, or drop the regex: four fields split by commas are easy to check by hand with `strings.Cut`.
- **`runtime.mallocgc` and `growslice` are at the top**: something allocates on every line. Look for `strings.Split` on the whole input, `fmt.Sprintf`, `make` and `copy` inside the loop.
- **Still quadratic after removing the regex**: there's a loop over every total for each line, and a full sort for each line. Use a `map[string]int` from name to index, and sort once at the end.
- **A test about cents fails after you drop `ParseFloat`**: build the cents from the digits yourself and skip the point. `999999999999.99` must give `99999999999999` exactly.
- **The allocation guard still fails by a little**: `strings.ToUpper` and `strings.TrimSpace` don't allocate when there's nothing to change, but a conversion like `string(bytes)` always does. Keep working with substrings of the input.
- **Benchmark numbers jump around by 20% between runs**: that's noise. Use `-count 6` or more and let `benchstat` judge. Never decide from one run.
- **Still stuck?** Open `solution/statement.go`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- **README**: this is a write-up project. Before and after `benchstat` tables, the flame graph before and after as screenshots, and a short list of each change with what it bought you ("removed the per-line regex compile: 30x").
- **Demo**: a blog post or a LinkedIn post: "How I made a statement parser 400x faster". Lead with the numbers, then the method: measure, profile, fix the biggest thing, measure again.
- **Tests and CI**: the allocation guards are your regression tests for performance. Run them in GitHub Actions with the badge. Explain why they count allocations instead of timing: CI machines are noisy, and a test that fails at random gets ignored.
- **Deploy link**: link the repository and the post. If you build the streaming stretch, a small CLI release is a nice touch.
- **What you'd do next**: say where the time goes now (likely the map lookups and the final sort), what you'd try next, and when you'd stop. Past a point, the database or the network costs more than this function, and the honest answer is "it's fast enough".
