# Kata: Running Percentile

Twende's dashboard shows the p95 latency of the booking API, live. Latencies arrive one at a time, and at any moment someone may ask for the current p95. Sorting everything again for each question is far too slow. Answer each question in O(log n).

Use the nearest-rank percentile from Day 77: with `n` samples sorted, the `p`-th percentile is the sample at position `⌈p × n / 100⌉`, counting from 1.

**Input**

- Line 1: `q p`, the number of operations and the percentile wanted (1 ≤ q ≤ 200 000, 1 ≤ p ≤ 100)
- Then `q` lines, each one of:
  - `add <ms>`: a request took this many milliseconds (0 ≤ ms ≤ 1 000 000 000)
  - `query`: what is the `p`-th percentile so far?

**Output**

For each `query`, one line: the percentile, or `-` if there are no samples yet.

**Example**

```
Input       Output
7 50        -
query       30
add 30      10
query       20
add 10
query
add 20
query
```

With samples 10 and 30, rank `⌈50 × 2 / 100⌉ = 1`, so the answer is 10. With 10, 20, 30 it's rank 2: 20.

**Limits:** 2 seconds per test case.

**Go tip:** keep two heaps. `low` is a max-heap holding the smallest `k` samples and `high` is a min-heap holding the rest, where `k = ⌈p × n / 100⌉`, computed in integers as `(p*n + 99) / 100`. The answer is the top of `low`. On each `add`, push into `low` if the value is no bigger than `low`'s top, otherwise into `high`, then move tops across until `low` holds exactly `k`. With `container/heap`, a max-heap is a min-heap whose `Less` says `>`.

Judge it from the repo root with `go run ./tools/kata 81`.
