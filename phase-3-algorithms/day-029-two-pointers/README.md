# Day 29: Two Pointers

## The brief

At Wakulima Market, two brothers price a long row of sacks. One starts at the left end, one at the right, and they walk towards each other. Neither ever walks back. Together they cover the row in one pass. That's the two pointers technique: two indexes into the same data, each moving in one direction only, so the whole job is O(n).

You'll use it four ways: checking a palindrome from both ends, finding two prices that add up to a budget in a sorted list, merging two sorted lists of M-Pesa transactions, and removing repeats in place.

```
IsPalindrome("Was it a rat I saw?")                 →  true
PairWithSum([]int{50, 100, 200, 450, 800}, 650)     →  2, 3, true      200 + 450
MergeSorted([]int{1, 4, 9}, []int{2, 4, 10})        →  [1 2 4 4 9 10]
Dedup([]int{20, 20, 50, 50, 50, 100})               →  [20 50 100]
```

## What you'll use

### Two pointers from both ends

Start `i` at 0 and `j` at the end. Move one of them each step until they meet. On a sorted slice, the sum `a[i] + a[j]` tells you which one to move:

```go
i, j := 0, len(a)-1
for i < j {
	sum := a[i] + a[j]
	if sum == target { return i, j, true }
	if sum < target { i++ } else { j-- }
}
```

Why is it safe to skip? If `a[i] + a[j]` is too small, then `a[i]` plus anything left of `j` is smaller still, so `a[i]` can't be in any pair. Drop it. Each step rules out one value, so at most n steps.

### Two pointers in the same direction

A read pointer `r` visits every element. A write pointer `w` marks where the next kept value goes. That's how `Dedup` works in place with no second slice. Merging uses one pointer per input, always taking the smaller front value.

Today's `Dedup` returns `sorted[:w]`. The result shares memory with the input ([Day 6](../../phase-1-terminal/day-006-chama-ledger/)), which is the point: no extra space.

### Spotting it in a problem

- The input is **sorted** (or sorting it first is fine) and you want **a pair** with some sum or difference.
- "palindrome", "reverse", "compare from both ends".
- "merge two sorted lists", "intersection of two sorted arrays".
- "in place", "O(1) extra space", "remove duplicates from a sorted array".

### Brute force vs optimal

| Problem | Brute force | Two pointers |
|---|---|---|
| pair with sum in a sorted list | every pair: O(n²) | both ends: O(n), no extra memory |
| merge two sorted lists | append and sort: O(n log n) | walk both: O(n) |
| remove duplicates | new slice plus a set: O(n) time, O(n) space | read and write pointers: O(n) time, O(1) space |
| almost a palindrome (kata) | try every deletion: O(n²) | one walk, two checks: O(n) |

[Day 28](../day-028-duplicate-receipts/)'s map solves Two Sum on unsorted data in O(n) but uses O(n) memory. If the data is already sorted, two pointers need none. If it isn't, sorting costs O(n log n). Knowing both, and saying why you chose one, is what interviewers listen for. See the patterns table in [reference/big-o.md](../../reference/big-o.md#patterns-and-their-complexity).

## Steps

1. Open `starter/pointers.go`. Write two small helpers: `alnum(c byte) bool` for ASCII letters and digits, and `lower(c byte) byte`, which adds `'a' - 'A'` to an upper-case letter.
2. Write `IsPalindrome`: `i` from the start, `j` from the end. Skip anything that isn't a letter or digit, compare the lower-case bytes, then move both in.
3. Write `PairWithSum` with the loop above.
4. Write `MergeSorted`: `make([]int, 0, len(a)+len(b))`, take the smaller front value while both have values left, then append whatever is left of each.
5. Write `Dedup`. Return early for an empty slice. Keep `w = 1`. For each `r` from 1, if `sorted[r]` differs from `sorted[w-1]`, copy it to `sorted[w]` and move `w` on. Return `sorted[:w]`.
6. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-029-two-pointers/starter
```

## The kata: Almost a Palindrome

Can a word become a palindrome by deleting at most one letter? This is LeetCode's "Valid Palindrome II" and a close cousin of HackerRank's "Palindrome Index". Walk in from both ends, and at the first mismatch try both deletions. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 29
```

## When you're stuck

- **`index out of range [-1]`** in `IsPalindrome("")`: `j` starts at -1. Loop with `for i < j`, which is false straight away.
- **`IsPalindrome("0P")` is true:** your lower-case conversion changed a digit. Only change bytes from `'A'` to `'Z'`.
- **`PairWithSum([]int{1, 5, 11}, 10)` returns true:** you used `i <= j`, so 5 was added to itself. Stop when the pointers meet.
- **`MergeSorted` loses the tail:** after the main loop one slice still has values. `append(out, a[i:]...)` and the same for `b`.
- **`MergeSorted(nil, nil)` returns `nil`:** a test asks for an empty slice. Start from `make([]int, 0, ...)`.
- **"`Dedup` made a new slice":** you appended to a fresh slice. Overwrite `sorted` and return `sorted[:w]`.
- **Still stuck?** Open `solution/pointers.go`, read it once, close it, then write your own version from memory.
