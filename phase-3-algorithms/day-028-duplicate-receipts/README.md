# Day 28: Duplicate Receipts

## The brief

Mama Njeri runs a busy duka in Githurai and keeps a list of every M-Pesa transaction code her till receives. A customer swears she paid twice. Somewhere in 200 000 codes is a repeat. Scanning the list once for each code is 2 × 10^10 comparisons. A map does it in one pass.

Today's tools are maps used as sets and as counters, the most useful trick in timed tests. Whenever a problem asks "have I seen this before?" or "how many of each?", reach for a map.

```
FirstDuplicate([]string{"QFT1", "RKD4", "SAB0", "RKD4", "QFT1"})   →  "RKD4", true
CountFrequencies([]string{"unga", "sukari", "unga"})               →  map[sukari:1 unga:2]
TopSeller([]string{"mandazi", "chai", "chai", "mandazi", "samosa"}) →  "chai", 2
TwoSum([]int{300, 700, 150, 850}, 1000)                            →  0, 1, true
```

`TwoSum` is the interview classic: two payments that together cover a Ksh1000 bill.

## What you'll use

### Maps as sets and counters

You met maps on Day 7. A set is a map whose values you ignore. `struct{}` takes no memory, so `map[string]struct{}` is the usual Go set:

```go
seen := map[string]struct{}{}
if _, ok := seen[code]; ok { /* repeat */ }
seen[code] = struct{}{}

counts := map[string]int{}
counts["unga"]++ // a missing key starts at 0, so this just works
```

A lookup, insert or delete is O(1) on average (see [reference/big-o.md](../../reference/big-o.md#common-go-operations)). `make(map[string]int, n)` sizes the map up front so it doesn't keep growing.

### Hashing a complement: Two Sum

For each amount `x`, the partner you need is `target - x`. Look it up in a map of the amounts you've already passed. If it's there, you're done. If not, remember `x` and move on. One pass, O(n).

```go
firstAt := map[int]int{} // amount → first index
for j, x := range amounts {
	if i, ok := firstAt[target-x]; ok { return i, j, true }
	...
}
```

Check before you store. That order stops `500` pairing with itself for a 1000 target, while `[500, 20, 500]` still finds positions 0 and 2.

### Ties and map order

Ranging over a map visits keys in a random order. If two items tie, decide the winner with a rule (here: alphabetical), not by whichever the loop met first. Otherwise your answer changes from run to run.

### Spotting it in a problem

- "duplicate", "first repeated", "unique", "distinct": a set.
- "how many times", "most frequent", "anagram": a counter.
- "two numbers that add up to", "pair with difference k": store what you've seen, look up the complement.
- "subarray with sum k" or "equal number of X and Y": prefix sums stored in a map (the kata).

### Brute force vs optimal

| Problem | Brute force | With a map |
|---|---|---|
| first duplicate | compare each code with every earlier one: O(n²) | set of seen codes: O(n) |
| most frequent | count each item by rescanning: O(n²) | one counting pass: O(n) |
| two sum | try every pair: O(n²) | complement lookup: O(n) |
| draws in a goal string (kata) | count every stretch: O(n²) | map of running scores: O(n) |

Sorting first is the other option: sort, then equal items sit next to each other. That's O(n log n) and needs no extra map, but it loses the original order, which `FirstDuplicate` and `TwoSum` need.

## Steps

1. Open `starter/receipts.go`. Write `FirstDuplicate` with a `map[string]struct{}`: for each code, return it if it's already in the set, otherwise add it.
2. Write `CountFrequencies`. Make the map with `make` so an empty list returns an empty map, not `nil`.
3. Write `TopSeller`: call `CountFrequencies`, then range over the counts. Replace the best when the count is higher, or when it's equal and the item is alphabetically smaller.
4. Write `TwoSum` with a map from amount to the first index it appeared at. Look up `target - x` before you store `x`, and only store an amount the first time you see it.
5. Run the tests from the repo root. Two of them use 200 000 items, so an O(n²) answer will be slow:

```bash
go test ./phase-3-algorithms/day-028-duplicate-receipts/starter
```

## The kata: Derby Draws

A string of Gor Mahia and AFC Leopards goals: count every stretch where the score is level. It looks like a counting-substrings problem, which is O(n²). A running score and a map make it O(n): the same prefix-sum idea from Day 27, with a map instead of an array. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 28
```

## When you're stuck

- **`assignment to entry in nil map`:** you declared `var seen map[string]struct{}` and wrote to it. Make it with `make(...)` or `map[string]struct{}{}`.
- **`TopSeller` passes sometimes and fails other times:** map order is random. Your tie-break must compare names, not rely on which item came first in the loop.
- **`TwoSum([]int{50}, 100)` returns true:** you stored `x` before looking up `target - x`, so 50 found itself. Look up first.
- **`TwoSum` returns `1, 2` for `[400, 400, 600]`:** you overwrote the index of `400` when you saw it again. Only store an amount the first time.
- **The big tests take many seconds:** there's a loop inside a loop somewhere. `slices.Contains` inside a loop counts: it's O(n) on its own.
- **The kata answer is too small:** start the map with `{0: 1}`. That counts draws that begin at the very first goal.
- **Still stuck?** Open `solution/receipts.go`, read it once, close it, then write your own version from memory.
