# Day 30: Sliding Window

## The brief

A boda boda rider in Kakamega wants to know his best week: the 7 days in a row that earned the most. The slow way adds up every 7-day block from scratch. The fast way keeps one running total, a window, and slides it: add the new day on the right, take away the day that falls off the left. One pass, whatever the window size.

Today you'll use fixed windows (best streak, moving averages of the price of unga) and windows that grow and shrink (the longest run with no repeat, the shortest run of chama contributions that reaches a target).

```
BestStreak([]int{300, 800, 100, 900, 200}, 2)          →  1100     900 + 200 ... or 800 + 100? No: 100 + 900 = 1000, 900 + 200 = 1100
MovingAverage([]int{180, 190, 200, 210}, 2)            →  [185 195 205]
LongestUniqueRun("ugaliugali")                         →  5        "ugali"
ShortestToReach([]int{200, 500, 100, 1000, 300}, 1200) →  2        1000 + 300
```

## What you'll use

### A fixed window

Add the first `k` values. Then for each new index `i`, add `a[i]` and subtract `a[i-k]`. Each step is O(1), so the whole scan is O(n), not O(n·k).

```go
sum += sales[i] - sales[i-k]
best = max(best, sum)
```

### A window that grows and shrinks

Two pointers moving the same way (Day 29). `right` moves forward every step and adds a value. While the window breaks the rule, `left` moves forward and removes one. Each index enters once and leaves once, so the inner `for` loop costs O(n) in total, not O(n) per step.

```go
for right := range a {
	add(a[right])
	for broken() {
		remove(a[left])
		left++
	}
	best = max(best, right-left+1)
}
```

For "no repeated byte", the rule is "this byte isn't in the window yet". An array `[256]int` of last positions is a fast map for bytes.

### Spotting it in a problem

- "**consecutive**", "**contiguous**", "**subarray**", "**substring**", "in a row", "streak".
- "of length k", "every window of k": a fixed window.
- "longest ... with at most ...", "shortest ... with at least ...": a window that grows and shrinks.
- It needs a rule that stays true when the window shrinks, such as "all values are positive". With negative numbers, "shortest subarray with sum ≥ target" isn't a plain sliding window any more.

### Brute force vs optimal

| Problem | Brute force | Sliding window |
|---|---|---|
| best k-day total | sum every window: O(n·k) | running total: O(n) |
| moving averages | same | O(n) |
| longest run, no repeats | check every substring: O(n²) or worse | O(n) |
| shortest run reaching a target | every start and end: O(n²) | O(n) |
| longest streak with ≤ k skips (kata) | every start: O(n²) | O(n) |

With n = 2 × 10^5 and k = 10^5, the brute force best streak is 2 × 10^10 additions: far over the 10^8 a second budget in [reference/big-o.md](../../reference/big-o.md#what-n-allows).

## Steps

1. Open `starter/window.go`. Write `BestStreak`: return 0 for a bad `k`, add up the first `k` values, then slide. Start `best` from the first window, not from 0, or an all-refund list gives the wrong answer.
2. Write `MovingAverage` the same way. Convert with `float64(sum) / float64(k)`; `sum / k` on ints drops the half. Return `[]float64{}` for a bad `k`.
3. Write `LongestUniqueRun` with a growing and shrinking window. Keep `var last [256]int`, where `last[c]` is one more than the index where byte `c` was last seen (so 0 means never). When `last[c] > left`, jump `left` to `last[c]`.
4. Write `ShortestToReach`: add `amounts[right]`, then while `sum >= target`, record the length and remove `amounts[left]`.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-030-sliding-window/starter
```

## The kata: Gym Streak

Find Kamau's longest run of gym days with at most `k` skipped days in it. It's the classic "longest subarray with at most k zeros", and the shape of a dozen other interview problems. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 30
```

## When you're stuck

- **`slice bounds out of range [:3] with capacity 2`:** you took `sales[:k]` before checking `k <= len(sales)`. Check first.
- **`BestStreak` returns 0 for all refunds:** you started `best` at 0. Start it at the first window's sum.
- **`MovingAverage` gives `[1 3]` for `[1 2 4]` and `k = 2`:** integer division. Convert to `float64` before dividing.
- **`LongestUniqueRun("abba")` is 3:** when you met the second `a`, you moved `left` back to just after the first `a`, which is before the current `left`. Only move `left` forward: `if last[c] > left`.
- **The big tests take seconds:** you're re-adding the whole window each step. Keep one running total.
- **The kata is off by one:** the window from `left` to `right` holds `right - left + 1` days.
- **Still stuck?** Open `solution/window.go`, read it once, close it, then write your own version from memory.
