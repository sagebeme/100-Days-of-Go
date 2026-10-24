# Day 27: How Fast Is Fast

## The brief

Phase 3 gets you ready for real timed coding tests, the HackerRank kind that companies send before an interview. Your code has to be right and it has to be fast enough. Today you learn to tell, before you write a line, whether an idea is fast enough.

Otieno runs a matatu SACCO. His app stores each day's takings for the year, and his manager keeps asking "how much did we make from day 40 to day 90?" Adding up 50 days is quick once. Adding them up for 200 000 questions is not. You'll build the trick that answers each question in one subtraction, a few other fast answers, and a benchmark that proves the difference.

```
SumTo(10)                        →  55          Wanjiru saves Ksh1 on day 1, Ksh2 on day 2 ... how much by day 10?
p := PrefixSums([]int{500, 200, 300})
p                                →  [0 500 700 1000]
RangeSum(p, 1, 2)                →  500         days 1 to 2: 200 + 300
KnockoutRounds(16)               →  4           16 school teams, 4 rounds to a champion
KnockoutRounds(17)               →  5
```

## What you'll use

### Big-O: counting steps, not seconds

Big-O describes how the number of steps grows with the size of the input, `n`. You count the loops, then keep only the biggest term and drop constant factors.

```go
total := 0
for _, x := range a { // runs n times
	total += x        // O(n)
}

for i := range a {         // n times
	for j := i + 1; j < len(a); j++ { // up to n times each
		_ = a[i] + a[j]    // about n²/2 steps: O(n²)
	}
}

for teams > 1 { teams = (teams + 1) / 2 } // halves each time: O(log n)
```

| Big-O | n = 2 × 10^5 means about | Name |
|---|---|---|
| O(1) | 1 step | constant: a formula, a map lookup |
| O(log n) | 18 steps | halving: binary search, knockout rounds |
| O(n) | 2 × 10^5 | one pass |
| O(n log n) | 3.6 × 10^6 | sorting |
| O(n²) | 4 × 10^10 | every pair |

### The 10^8 rule

A judge does roughly 10^8 simple steps per second. Put the biggest `n` from the constraints into your Big-O. Under 10^8: safe. Around 10^9: it probably times out. 4 × 10^10: it certainly does. The table in [reference/big-o.md](../../reference/big-o.md#what-n-allows) turns this around: from the largest `n`, read off the slowest algorithm you can get away with. Keep that page open for all of Phase 3.

### Prefix sums

Build one array where `p[i]` is the total of the first `i` values. Then the sum of `a[l..r]` is `p[r+1] - p[l]`: everything up to `r`, minus everything before `l`. Building `p` is O(n), once. Each question is then O(1).

```
a = [500 200 300]
p = [0 500 700 1000]
sum of a[1..2] = p[3] - p[1] = 1000 - 500 = 500
```

### Benchmarks with testing.B

**New today.** A function named `BenchmarkXxx(b *testing.B)` in a `_test.go` file is a benchmark. `go test -bench .` runs it many times and reports the time per run. `b.Loop()` (Go 1.24 and later) runs the loop the right number of times and starts the clock on its first call, so setup before the loop isn't counted.

```go
func BenchmarkRangeSumPrefix(b *testing.B) {
	p := PrefixSums(data) // setup: not timed
	for b.Loop() {
		RangeSum(p, 10, 99_990)
	}
}
```

```
BenchmarkRangeSumNaive-4     76086 ns/op
BenchmarkRangeSumPrefix-4    2.3 ns/op
```

Same answer, about 30 000 times faster. `ns/op` is nanoseconds per run.

### Fast stdin and stdout

On a judge, slow input and output can time out a correct O(n) program. Wrap stdin in a `bufio.Reader` once, wrap stdout in a `bufio.Writer`, and always `Flush`:

```go
in := bufio.NewReader(os.Stdin)
out := bufio.NewWriter(os.Stdout)
defer out.Flush() // without this, nothing is printed
fmt.Fscan(in, &n)
fmt.Fprintln(out, answer)
```

For 10^6 numbers, use `bufio.Scanner` with `bufio.ScanWords` and `strconv.Atoi`. [reference/big-o.md](../../reference/big-o.md#fast-input-and-output) has the full template and a timing table. Copy it into every kata from now on.

### Spotting it in a problem

- **"n ≤ 2 × 10^5" or bigger:** O(n²) is out. Look for O(n) or O(n log n).
- **"q queries", each about a range:** precompute something (prefix sums) so each query is O(1) or O(log n).
- **"n ≤ 10^18":** you can't even loop to n. You need a formula or O(log n).
- **"print the answer modulo 10^9 + 7" or answers above 2 × 10^9:** use 64-bit `int` and watch for overflow.

### Brute force vs optimal

| Question | Brute force | Optimal |
|---|---|---|
| 1 + 2 + ... + n | loop: O(n) | n(n+1)/2: O(1) |
| sum of `a[l..r]`, q times | loop each time: O(n·q) | prefix sums: O(n + q) |
| rounds in a knockout | simulate every match: O(n) | halve the count: O(log n) |
| pairs D before W (the kata) | check every pair: O(n²) | one pass with a counter: O(n) |

Write the brute force first when you're unsure. It is easy to get right and it checks your fast version on small inputs. Then read the constraints and ask if it's fast enough.

## Why it matters at work

A report that loops over every transaction for every customer works on the test data and falls over on the first big client. In code review, "this is O(n²) on the number of orders" is a common and serious comment, and a benchmark is how you back up a claim that a change is faster. Go makes benchmarks part of the test files, so the proof sits next to the code and anyone can rerun it with `go test -bench`.

## Steps

1. Open `starter/fast.go`. Write `SumTo` with Gauss's formula `n * (n + 1) / 2`. Return 0 when `n < 1`. One test uses n = 10^9: a loop would pass but slowly, and the formula is instant.
2. Write `PrefixSums`: make a slice of `len(a)+1` zeros, then `p[i+1] = p[i] + a[i]` in a loop.
3. Write `RangeSum` as one line: `prefix[r+1] - prefix[l]`. Draw the example above on paper if the `+1` confuses you.
4. Write `KnockoutRounds`: while more than one team is left, `teams = (teams + 1) / 2` and count a round. The `+ 1` rounds up, which gives the bye. Count the steps: the loop runs about log₂(n) times.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-027-how-fast-is-fast/starter
```

6. Run the benchmarks and compare the two lines:

```bash
go test -bench . ./phase-3-algorithms/day-027-how-fast-is-fast/starter
```

7. Write your own benchmark in `starter/fast_test.go`: `BenchmarkSumToLoop` that adds 1 to 1 000 000 with a loop, next to `BenchmarkSumTo` that calls your formula. Run it again.

## The kata: Deposit Pairs

Count every pair in an M-Pesa agent's log where a deposit came before a withdrawal. The log is 200 000 characters long, so checking every pair is too slow: you need one pass. This is the pattern for the whole phase: read the constraints, work out the budget, then pick the algorithm. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 27
```

## When you're stuck

- **`SumTo(1000000000)` is wrong but small ones are right:** you divided before multiplying, `n / 2 * (n + 1)`, which drops the half when `n` is odd. Multiply first: `n * (n + 1) / 2`.
- **`index out of range [3] with length 3` in `PrefixSums`:** the result needs one more slot than the input. `make([]int, len(a)+1)`.
- **`RangeSum` is off by one value:** it's `prefix[r+1] - prefix[l]`, not `prefix[r] - prefix[l]`. `prefix[r]` stops just before `a[r]`.
- **`KnockoutRounds(17)` gives 4:** `teams / 2` rounds down and loses the team with a bye. Use `(teams + 1) / 2`.
- **`go test -bench .` prints only `PASS`:** the pattern after `-bench` is a regular expression for benchmark names. `.` matches all of them. Check the function starts with `Benchmark` and takes `*testing.B`.
- **The kata times out on the last case:** you're checking every pair. Keep a count of `D`s seen so far. Each `W` adds that count to the answer.
- **The kata prints nothing on the big case with a `bufio.Scanner`:** the line is longer than the 64 KB default and `Scan` returned false. Use `fmt.Fscan` on a `bufio.Reader`, or call `sc.Buffer(make([]byte, 1<<20), 1<<20)`.
- **Still stuck?** Open `solution/fast.go`, read it once, close it, then write your own version from memory.
