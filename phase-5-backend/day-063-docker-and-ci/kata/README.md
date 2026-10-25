# Kata: Build Order

In a GitHub Actions workflow, `needs: test` means a job waits until `test` has finished. Given every job and what it needs, print the order a single runner would use: a job runs only after everything it needs, and when several jobs are ready at once, the alphabetically smallest goes first. If the needs go round in a circle, no order exists.

**Input**

- Line 1: `n m`, the number of jobs and of needs (1 ≤ n ≤ 100 000, 0 ≤ m ≤ 200 000)
- Line 2: the `n` job names, separated by spaces, all different
- Next `m` lines: `job need`, meaning `job` needs `need` to run first. Both are names from line 2, and no pair appears twice.

**Output**

One line: all `n` job names in run order, separated by spaces. Or `CYCLE` if there's no valid order.

**Example**

```
Input
6 6
test lint build docker deploy smoke
build test
build lint
docker build
smoke docker
deploy smoke
deploy docker

Output
lint test build docker smoke deploy
```

`lint` and `test` need nothing, and `lint` comes first alphabetically. After both, `build` is ready, then `docker`, then `smoke`, and `deploy` needs both of those.

**Limits:** 2 seconds per test case. "Alphabetically" means comparing the names as Go strings with `<`.

**Go tip:** this is a topological sort. Count, for each job, how many needs it still waits for, and keep the ready ones (count 0) in a min-heap so you can always take the smallest. `container/heap` needs a type with `Len`, `Less`, `Swap`, `Push` and `Pop`. If the heap runs dry before every job has run, the rest are stuck in a cycle.

Judge it from the repo root with `go run ./tools/kata 63`.
