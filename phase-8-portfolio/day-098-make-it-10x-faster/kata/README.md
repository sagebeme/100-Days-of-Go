# Kata: Exact Budget

A boda rider records his net takings for each day: money in minus fuel and repairs, so a day can be negative. He wants to know how many runs of consecutive days netted exactly `K` cents. The obvious answer checks every run, which is O(N²), and with 200 000 days that's 2×10^10 steps. Make it 10x faster than that. Much more than 10x, in fact.

**Input**

- Line 1: `N` and `K` (1 ≤ N ≤ 200 000, |K| ≤ 10^15)
- Line 2: `N` integers, the daily totals in cents (|total| ≤ 10^9)

**Output**

One line: the number of pairs `(i, j)` with `i ≤ j` such that the totals of days `i` to `j` add up to exactly `K`. The answer can be bigger than 2^31, so use `int64`.

**Example**

```
Input                            Output
7 500                            8
200 300 -500 500 0 250 250
```

Some of the eight: `200 300`, `500`, `500 0`, `250 250`, and `200 300 -500 500`.

**Limits:** 2 seconds per test case.

**Go tip**

Keep a running total, the prefix sum. The days after position `i` up to `j` add up to `prefix[j] - prefix[i]`. So at each day, the number of runs ending today that hit `K` is the number of earlier prefix sums equal to `prefix - K`. A map counts them:

```go
seen := map[int64]int64{0: 1} // the empty prefix, before day 1
prefix += x
count += seen[prefix-k]
seen[prefix]++
```

That's O(N). It's LeetCode's "subarray sum equals K", and the same move as today's main exercise: replace the inner loop with a lookup.

Judge it from the repo root with `go run ./tools/kata 98`.
