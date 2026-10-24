# Kata: Latency Percentiles

Your logger now records how long every request took. An average hides the slow ones, so teams watch percentiles instead: p50 is the typical request, p99 is the slowest one in a hundred, the passenger whose booking spinner never stops.

Use the **nearest-rank** method. Sort the `n` durations from smallest to largest. For a percentile `p`, the rank is `p × n / 100` rounded **up** to a whole number, and the answer is the duration at that rank, counting from 1.

**Input**

- Line 1: `n`, the number of requests (1 ≤ n ≤ 20 000)
- Line 2: `n` durations in milliseconds (1 ≤ d ≤ 1 000 000 000)
- Line 3: `q`, the number of percentiles asked for (1 ≤ q ≤ 100)
- Line 4: `q` whole numbers `p` (1 ≤ p ≤ 100)

**Output**

One line per percentile, in the order asked: `p` then the percentile, then a space and the duration, like `p99 1203`.

**Example**

```
Input                                Output
10                                   p50 15
12 40 8 3 1203 30 31 15 22 9         p90 40
4                                    p99 1203
50 90 99 100                         p100 1203
```

Sorted: `3 8 9 12 15 22 30 31 40 1203`. For p50 the rank is 5, which is 15. For p99 it is 9.9, rounded up to 10, which is 1203. One slow request owns p99.

**Limits:** 2 seconds per test case.

**Go tip**

Round up with whole numbers, not floats: `(p*n + 99) / 100` is `p × n / 100` rounded up. Sort once with `sort.Ints`, then answer every query by indexing, so the work is O(n log n + q).

```go
rank := (p*n + 99) / 100 // 1-based
fmt.Fprintf(out, "p%d %d\n", p, sorted[rank-1])
```

Judge it from the repo root with `go run ./tools/kata 54`.
