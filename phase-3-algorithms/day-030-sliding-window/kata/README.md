# Kata: Gym Streak

Kamau logs his gym days in Kasarani as a string: `G` for a day he trained and `.` for a day he skipped. He wants to brag about his longest streak, and he allows himself up to `k` skipped days inside it. A streak is a run of consecutive days. Find the longest one with at most `k` dots in it.

**Input**

- Line 1: `n` and `k` (1 ≤ n ≤ 2 × 10^5, 0 ≤ k ≤ n)
- Line 2: the log, `n` characters, each `G` or `.`

**Output**

One line: the length of the longest run of consecutive days containing at most `k` skipped days.

**Example**

```
Input           Output
11 1            5
GG.G..GGG.G
```

With one skip allowed, the best runs are `GG.G` (4 days) and `GGG.G` (5 days, the last five).

**Limits:** 2 seconds per test case.

**Read the constraints.** Trying every start and extending to the right is O(n²), about 2 × 10^10 steps when `k` is large. A sliding window is O(n): move the right end one day at a time, and whenever the window holds more than `k` dots, move the left end until it doesn't.

**Go tip**

`max` is built in since Go 1.21: `best = max(best, right-left+1)`.

Judge it from the repo root with `go run ./tools/kata 30`.
