# Day 31: Sorting Showdown

## The brief

The Football Kenya Federation's intern prints the Premier League top scorers table, and it's in the wrong order. Today you'll write the two sorts every interviewer expects you to know, merge sort and quicksort, then sort the table with a custom rule: most goals first, then by name.

You'll rarely write a sort at work: `slices.Sort` is faster and tested. But the ideas inside them, splitting a problem in half and merging, or partitioning around a pivot, come up in problem after problem, like today's kata.

```
MergeSort([]int{5, 2, 9, 1})     →  [1 2 5 9]       a new slice; the input is unchanged
QuickSort(a)                     →  a is now sorted in place
SortScorers(table)               →  Baraka 15, Wekesa 15, Achieng 12, Otieno 12, Mwangi 9, ...
```

## What you'll use

### Merge sort: O(n log n) always

Split the slice in half, sort each half (recursively), then merge the two sorted halves, which you wrote on Day 29. There are about log₂ n levels of halving and each level merges n values in total, so the cost is O(n log n) whatever the input. It is **stable**: equal values keep their order, as long as the merge takes from the left on a tie (`<=`). It needs O(n) extra memory for the merged slices.

### Quicksort: O(n log n) on average, in place

Pick a pivot. Move everything smaller to its left and everything bigger to its right (that's the **partition**). Then sort the two sides. No extra slice is needed.

The trap: if the pivot is always the smallest value, one side is empty and the other has n−1 values, so it's O(n²). Taking the first element as the pivot does exactly that on sorted input. Take the middle element. The simple Lomuto partition has a second trap: on 200 000 equal values it is O(n²) too. Hoare's partition below moves both ends inwards and splits equal values down the middle:

```go
pivot := a[len(a)/2]
i, j := 0, len(a)-1
for i <= j {
	for a[i] < pivot { i++ }
	for a[j] > pivot { j-- }
	if i <= j { a[i], a[j] = a[j], a[i]; i++; j-- }
}
// a[:j+1] <= pivot, a[i:] >= pivot: sort those two parts
```

The tests include 200 000 sorted values and 200 000 equal values to catch both traps. Go's own `slices.Sort` uses pattern-defeating quicksort, which avoids them for you.

### Custom comparators

`slices.SortFunc(s, cmp)` sorts with your rule. `cmp(x, y)` returns a negative number if `x` goes first, positive if `y` goes first, and 0 if they tie. `cmp.Compare` from the `cmp` package does this for numbers and strings. Swap the arguments to sort descending. Check one key, and only on a tie move to the next:

```go
slices.SortFunc(table, func(x, y Scorer) int {
	if c := cmp.Compare(y.Goals, x.Goals); c != 0 { // y, x: most goals first
		return c
	}
	return cmp.Compare(x.Name, y.Name)
})
```

`sort.Slice(s, func(i, j int) bool {...})` from Day 10 works too. Neither is stable. Use `slices.SortStableFunc` when the original order of ties matters.

### Spotting it in a problem

- "**k-th** smallest / largest", "**median**", "closest pair", "minimum difference": sort, then scan neighbours.
- "**merge** intervals", "meeting rooms", "schedule": sort by start or end time.
- "order by X, then by Y": a custom comparator.
- "count pairs where a bigger value comes first", "minimum adjacent swaps": count inversions during merge sort (the kata).

### Brute force vs optimal

| Problem | Brute force | Better |
|---|---|---|
| sort n values | bubble or insertion sort: O(n²) | merge sort, quicksort, `slices.Sort`: O(n log n) |
| minimum difference between any two | every pair: O(n²) | sort, compare neighbours: O(n log n) |
| count inversions (kata) | every pair: O(n²) | count during merge sort: O(n log n) |

The 10^8 rule from [reference/big-o.md](../../reference/big-o.md#what-n-allows): n = 2 × 10^5 needs O(n log n), which is about 3.6 × 10^6 steps.

## Steps

1. Open `starter/sorting.go`. Write `MergeSort`. For `len(a) <= 1`, return `slices.Clone(a)` so the caller never shares memory with the input. Otherwise sort `a[:mid]` and `a[mid:]` and merge them into a new slice, taking from the left on a tie.
2. Write `QuickSort` with the Hoare partition above. Then sort `a[:j+1]` and `a[i:]` by calling `QuickSort` on each.
3. Run the tests. If the big cases are slow, check your pivot.
4. Improve `QuickSort`: recurse only into the smaller part and loop on the bigger one by setting `a` to it. Now the recursion depth stays near log n even in a bad case.
5. Write `SortScorers` with `slices.SortFunc` and `cmp.Compare`. Import `"cmp"` and `"slices"`.
6. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-031-sorting-showdown/starter
```

## The kata: Line Up by Marks

How many swaps of neighbours does it take to line pupils up by exam mark? Count the inversions while you merge sort. This is the classic proof that merge sort is a technique, not just a sort. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 31
```

## When you're stuck

- **`runtime: goroutine stack exceeds 1000000000-byte limit`:** your recursion never stops. `MergeSort` must return for `len(a) <= 1`, and `QuickSort` must loop only while `len(a) > 1`.
- **The tests take 20 seconds or more:** your quicksort is O(n²) on the big sorted or all-equal cases. Use the middle element as the pivot and Hoare's partition.
- **"`MergeSort` changed its input":** you sorted `a` itself. Build the result in a new slice.
- **`QuickSort` loops forever on `[2, 1]`:** in the partition, `i++` and `j--` must happen after every swap, or the same pair swaps forever.
- **Ties come out as `Wekesa` before `Baraka`:** your comparator returns after the goals check even when goals are equal. Return only if `c != 0`.
- **The kata answer is too big:** equal marks aren't inversions. Take from the left half when `left[i] <= right[j]`.
- **Still stuck?** Open `solution/sorting.go`, read it once, close it, then write your own version from memory.
