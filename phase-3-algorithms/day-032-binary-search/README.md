# Day 32: Binary Search

## The brief

Guess a number between 1 and 1 000 000, and you're told "higher" or "lower" each time. Twenty guesses always do it, because each one halves what's left. You played this on Day 12. Today it becomes a tool: find a price in a sorted list of a million M-Pesa transactions in 20 steps, count how many transactions fall in a price band, and, the part that wins timed tests, binary search over possible **answers** instead of over data.

```
Search([]int{50, 100, 200, 500, 1000}, 200)               →  2
FirstAtLeast([]int{50, 100, 100, 500}, 100)               →  1      the first 100
CountInRange([]int{50, 100, 100, 500, 1000}, 100, 500)    →  3
IntSqrt(26)                                               →  5      5*5 = 25 <= 26 < 36
```

## What you'll use

### The loop and its invariant

Keep a range `[lo, hi]` where the answer must be. Look at the middle. Throw away the half that can't hold it. Stop when the range is empty or one value.

```go
lo, hi := 0, len(a)-1
for lo <= hi {
	mid := lo + (hi-lo)/2 // not (lo+hi)/2: that can overflow on huge ranges
	if a[mid] == x { return mid }
	if a[mid] < x { lo = mid + 1 } else { hi = mid - 1 }
}
return -1
```

Most binary search bugs are off-by-one errors. Say the invariant out loud ("the answer is in `[lo, hi]`") and check each update keeps it true.

### Lower bound: the first value at least x

`FirstAtLeast` searches `[0, len(a)]`. When `a[mid] >= x`, `mid` could be the answer, so `hi = mid` keeps it. Otherwise `lo = mid + 1`. It stops when `lo == hi`. This one function handles repeats, missing values and "count in a range": `FirstAtLeast(a, hi+1) - FirstAtLeast(a, lo)`.

The standard library has it as `sort.Search(n, func(i int) bool { return a[i] >= x })` and `slices.BinarySearch`. Write it yourself today; use those later.

### Binary search on the answer

If you can check "does answer `c` work?" and the answers that work are all on one side (if 6 kg works, 7 kg works), the answer is searchable even with no sorted array in sight:

```go
lo, hi := smallestPossible, largestPossible
for lo < hi {
	mid := lo + (hi-lo)/2
	if works(mid) { hi = mid } else { lo = mid + 1 }
}
// lo is the smallest answer that works
```

The cost is O(log(range) × cost of `works`). A range of 10^9 is 30 steps. `IntSqrt` is the simplest case: "does `r*r <= n`?" is true for small `r` and false after. Check it as `r <= n/r` so `r*r` can't overflow.

### Spotting it in a problem

- The input is **sorted**, or "find the first / last position of x".
- "**minimum** capacity / speed / time such that ...", "**maximum** distance such that ...". The words "at least", "at most" and "smallest possible" next to a yes/no check are the signal.
- The answer range is huge (10^9 or 10^18) but checking one answer is cheap.
- Many queries on fixed data: sort once, then binary search each query.

### Brute force vs optimal

| Problem | Brute force | Binary search |
|---|---|---|
| find x in sorted data | scan: O(n) | O(log n) |
| q lookups in n values | O(n·q) | sort once, then O((n + q) log n) |
| integer square root of 10^18 | try every r: O(√n) = 10^9 | O(log n) = 60 steps |
| smallest boda capacity (kata) | try every capacity: O(n × range) | O(n log range) |

See the "n ≤ 10^18" row of [reference/big-o.md](../../reference/big-o.md#what-n-allows): when even a loop to n is too slow, you need O(log n).

## Steps

1. Open `starter/search.go`. Write `Search` with the loop above. Don't import `sort` or `slices`.
2. Write `FirstAtLeast` with `lo, hi := 0, len(sorted)` and `for lo < hi`. Work through `[50, 100, 100, 500]` with `x = 100` on paper before you run it.
3. Write `CountInRange` as two calls to `FirstAtLeast`. Return 0 when `lo > hi`.
4. Write `IntSqrt`. The answer is in `[0, 10^9]`. Use `mid := lo + (hi-lo+1)/2` (round up), and `lo = mid` when `mid <= n/mid`, else `hi = mid - 1`.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-032-binary-search/starter
```

## The kata: Boda Deliveries

The smallest bike capacity that delivers every parcel within `d` days. You can't sort the parcels and there's no array to search. You binary search the answer. LeetCode calls this "Capacity To Ship Packages Within D Days", and the same shape appears as "Koko Eating Bananas" and "split array largest sum". Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 32
```

## When you're stuck

- **The test hangs:** your loop never shrinks the range. With `for lo < hi` and `lo = mid`, round `mid` up: `lo + (hi-lo+1)/2`. With `hi = mid`, round down.
- **`index out of range [5] with length 5`** in `FirstAtLeast`: `hi` starts at `len(sorted)`, which isn't an index. Only read `sorted[mid]`, and `mid < hi` always holds inside the loop.
- **`FirstAtLeast(prices, 100)` returns 2:** you stopped at any 100, not the first. On `>=`, set `hi = mid` and keep going.
- **`IntSqrt(10^18)` is wrong or negative:** `mid*mid` overflowed. Compare `mid <= n/mid`.
- **The kata is too slow:** you're trying capacities one by one. Search between the heaviest parcel and the total.
- **The kata answer is too small:** your check let a day go over capacity. Start a new day when `load + w > capacity`, before adding `w`.
- **Still stuck?** Open `solution/search.go`, read it once, close it, then write your own version from memory.
