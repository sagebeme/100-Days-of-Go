# Kata: Polish Plan

The last kata of the course. You have `H` hours before applications close, and a list of polishing tasks across your projects: a README here, a deploy there, a demo video, a CI badge. Each task takes some hours and has some impact on how your portfolio reads. Each task can be done once or not at all. Pick the tasks that give the most impact within `H` hours.

**Input**

- Line 1: `N` and `H`, the number of tasks and the hours you have (1 ≤ N ≤ 1 000, 0 ≤ H ≤ 100 000)
- The next `N` lines: `hours impact` for one task (1 ≤ hours ≤ 100 000, 0 ≤ impact ≤ 10^6)

**Output**

- Line 1: the most total impact you can get using at most `H` hours
- Line 2: the fewest hours that reach that impact (the first line's plan might not need all `H`)

**Example**

```
Input       Output
4 10        50
6 30        10
4 20
5 25
3 10
```

The 6-hour and 4-hour tasks together give 50 in exactly 10 hours. Nothing else reaches 50, so the second line is 10.

**Limits:** 2 seconds per test case.

**Go tip**

This is the 0/1 knapsack, the most common dynamic programming question in interviews (Day 39). Keep one slice, `best[c]` for every budget `c`, and for each task go through `c` from high to low, so the task can't be counted twice:

```go
for c := h; c >= hours; c-- {
	best[c] = max(best[c], best[c-hours]+impact)
}
```

That's O(N × H). Use `int64` for impact: 1 000 tasks of 10^6 each is 10^9, too close to the edge of a 32-bit int.

Judge it from the repo root with `go run ./tools/kata 100`. Then go and do the real thing.
