# Kata: Coverage Report

`go test -cover` tells you what share of your code the tests ran. Build a tiny version of it. A file has `L` lines, numbered from 1. Each test ran one block of lines, from `l` to `r` inclusive. Blocks can overlap. Work out how many lines at least one test ran, the percentage, and which lines no test touched.

**Input**

- Line 1: `L` and `n`, the number of lines and the number of blocks (1 ≤ L ≤ 1 000 000, 0 ≤ n ≤ 100 000)
- The next `n` lines: `l` and `r` (1 ≤ l ≤ r ≤ L)

**Output**

- Line 1: `C/L lines (P%)`, where `C` is the number of covered lines and `P` is `100 × C / L` **rounded down** to one decimal place
- Line 2: the uncovered lines as ranges, in order, separated by spaces. A range of one line is just the number (`9`); a longer one is `first-last` (`3-5`). If every line is covered, print `none`.

**Example**

```
Input       Output
10 3        6/10 lines (60.0%)
1 2         3 5 9-10
6 8
4 4
```

Lines 1, 2, 4, 6, 7 and 8 ran. Line 3, line 5, and lines 9 to 10 didn't.

**Limits:** 2 seconds per test case.

**Go tip**

Marking every line of every block could mean 10^11 steps. Two fast ways:

- A **difference array**: add 1 at `l` and subtract 1 at `r+1`, then take a running sum. A line is covered when the sum is above 0.
- **Sort and merge**: sort the blocks by `l` and join each one to the previous when they overlap or touch.

To round down to one decimal without floats, work in tenths of a percent: `p := c * 1000 / L`, then print `p/10` and `p%10`.

Judge it from the repo root with `go run ./tools/kata 19`.
