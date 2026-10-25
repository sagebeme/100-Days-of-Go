# Day 39: Dynamic Programming

## The brief

Day 35's recursion tried every possibility. That's fine until the same smaller problem comes up again and again. Counting the ways up the 90 steps of a footbridge in Mombasa, one or two at a time, plain recursion asks "how many ways from step 45?" billions of times. Dynamic programming (DP) asks it once and writes the answer down.

You'll do it two ways, memoisation and tabulation, on four classics: the stairs, the fewest M-Pesa agent notes for an amount (where Day 38's greedy breaks), the longest common subsequence of two names, and the best paying load for a boda boda.

```
ClimbWays(4)                                              →  5      1111, 112, 121, 211, 22
FewestNotes([]int{1, 3, 4}, 6)                            →  2      3 + 3, not greedy's 4 + 1 + 1
FewestNotes([]int{200, 500}, 300)                         →  -1     can't be done
LCS("KAMAU", "KAMUNYA")                                   →  4      "KAMU"
BestLoad([]int{10, 20, 30}, []int{60, 100, 120}, 50)      →  220    the 20 kg and 30 kg parcels
```

## What you'll use

### The two questions

1. **What is a state?** The smallest description of a subproblem: "the ways from step i", "the fewest notes for amount x", "the LCS of the first i bytes of a and the first j of b".
2. **How does a state use smaller ones?** The recurrence. For the stairs: the last move was 1 step or 2, so `ways(n) = ways(n-1) + ways(n-2)`.

The cost is the number of states times the work per state. That's the DP row in [reference/big-o.md](../../reference/big-o.md#patterns-and-their-complexity).

### Memoisation: top down

Write the recursion, then add a map. Check it before computing and store the answer after:

```go
if w, ok := memo[n]; ok {
	return w
}
w := ways(n-1) + ways(n-2)
memo[n] = w
```

Without the map, `ways(90)` makes about 10^19 calls. With it, 91. It's the easiest DP to write because it's the recursion you'd write anyway.

### Tabulation: bottom up

Fill a slice from the smallest state up, so every answer you need is already there. No recursion, so no stack depth to worry about, and often less memory:

```go
best := make([]int, amount+1) // best[0] = 0: no notes for nothing
for x := 1; x <= amount; x++ {
	// best[x] = 1 + the smallest best[x-n] over every note n <= x
}
```

### Two-dimensional tables

LCS has two indexes, so it has a 2D table. `t[i][j]` is the answer for `a[:i]` and `b[:j]`. If `a[i-1] == b[j-1]`, use that byte: `t[i-1][j-1] + 1`. Otherwise drop a byte from one string: `max(t[i-1][j], t[i][j-1])`. Make it with a slice of slices, each row its own `make`.

### 0/1 knapsack in one row

`best[c]` is the most you can earn with capacity `c` using the parcels seen so far. For each parcel, update the row **from high capacity down**, so `best[c-w]` still means "without this parcel" and no parcel is taken twice. Going up instead lets each parcel be taken many times, which is the right loop for `FewestNotes`, where notes repeat.

### Spotting it in a problem

- "**how many ways**", "number of paths", "count the decodings": add up smaller counts.
- "**minimum cost** / fewest / maximum profit" where greedy has a counter-example.
- "longest common / increasing subsequence", "edit distance", "knapsack", "coin change".
- A recursive brute force where the same arguments show up again. Print the arguments, and if they repeat, memoise.
- The answer is "modulo 10^9 + 7": almost always a counting DP with a huge answer.

### Brute force vs optimal

| Problem | Brute force | DP |
|---|---|---|
| stairs, n = 90 | recursion: about 1.6^n calls | n states: O(n) |
| fewest notes | try every combination: exponential | amount × notes: O(amount · k) |
| LCS of lengths n and m | every subsequence of one: O(2^n · m) | O(n · m) |
| knapsack, n parcels | every subset: O(2^n) | O(n · capacity) |
| decode the SMS (kata) | every split: exponential | O(n) |

The knapsack DP is fast only when `capacity` is small. With n = 100 parcels and a capacity of 10^4 it's 10^6 steps. With a capacity of 10^12 it's useless. Read both numbers in the constraints.

## Steps

1. Open `starter/dp.go`. Write `ClimbWays` with memoisation: a `memo := map[int]int{}` and a recursive closure `ways`. Base cases: a negative `n` is 0 ways and `n == 0` is 1.
2. Write `FewestNotes` by tabulation. Use -1 in the table for "impossible", and skip `best[x-n]` when it's -1. `best[0]` is 0.
3. Write `LCS` with a `(len(a)+1) × (len(b)+1)` table. Row 0 and column 0 stay 0: the LCS with an empty string is empty.
4. Write `BestLoad` with one row of `capacity+1` ints, looping each parcel's capacities from `capacity` down to its weight.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-039-dynamic-programming/starter
```

## The kata: Decode the SMS

Count the messages a string of digits could spell when `A` is `1` and `Z` is `26`. It's LeetCode's "Decode Ways": a one-dimensional DP, plus the zeros, which are where most answers go wrong. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 39
```

## When you're stuck

- **`ClimbWays(90)` never finishes:** the memo isn't being used. Check it before recursing and store the answer before returning. Make the map outside the closure, or each call gets a fresh, empty one.
- **`FewestNotes([]int{2, 5}, 3)` returns a huge or negative number:** you added 1 to an "impossible" entry. Skip any `best[x-n]` that is -1.
- **`FewestNotes` returns 3 for `[1, 3, 4]` and 6:** you took the biggest note first. That's greedy. Try every note at every amount.
- **`index out of range [5] with length 5`** in `LCS`: the table needs `len(a)+1` rows and `len(b)+1` columns. Compare `a[i-1]` and `b[j-1]`.
- **`BestLoad` takes the 10 kg parcel ten times:** your capacity loop goes up. Go down, from `capacity` to `w`.
- **The kata says `06` has 1 way:** a `0` is never a letter on its own. When `s[i]` is `'0'`, the ways from `i` are 0.
- **Still stuck?** Open `solution/dp.go`, read it once, close it, then write your own version from memory.
