# Kata: Line Up by Marks

At a school assembly in Nyeri, pupils stand in one line with their exam marks. The head teacher wants them in order from the lowest mark to the highest, and the only move allowed is swapping two pupils standing next to each other. How many swaps does it take?

That number is exactly the number of **inversions**: pairs of pupils where the one in front has a higher mark than the one behind. Each swap of neighbours in the wrong order fixes one inversion.

**Input**

- Line 1: `n`, the number of pupils (1 ≤ n ≤ 2 × 10^5)
- Line 2: `n` marks, separated by spaces (0 ≤ mark ≤ 100)

**Output**

One line: the number of pairs (i, j) with i < j and mark[i] > mark[j]. Equal marks are not an inversion.

**Example**

```
Input           Output
5               4
2 1 3 1 2
```

The pairs are (2, 1), (2, 1), (3, 1) and (3, 2), counting by position.

**Limits:** 2 seconds per test case.

**Read the constraints.** Checking every pair is O(n²): 2 × 10^10 steps for the largest n. Count while you merge sort, O(n log n). When you take a value from the right half, it jumps ahead of every value still waiting in the left half, and each of those is an inversion: add `len(left) - i`. The answer can reach about 2 × 10^10, so use `int`.

**Go tip**

This is HackerRank's "Merge Sort: Counting Inversions". Write `sortCount(a []int) int` that sorts `a` and returns the count, built from two recursive calls plus a merge. Read the numbers with `bufio.Scanner` and `bufio.ScanWords` (see [reference/big-o.md](../../../reference/big-o.md#reading-bufioscanner-with-a-bigger-buffer)).

Judge it from the repo root with `go run ./tools/kata 31`.
