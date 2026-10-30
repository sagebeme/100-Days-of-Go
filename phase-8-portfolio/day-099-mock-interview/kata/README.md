# Kata: Fewest Changes Fare

The last problem of the mock interview, and the hardest. A traveller wants the cheapest way across a made-up network of stages, but each change of matatu wastes time, so they'll accept at most `K` changes: at most `K + 1` rides. Rides go one way only.

**Input**

- Line 1: `N`, `M` and `K`: stages numbered 1 to `N`, `M` rides, at most `K` changes (1 ≤ N ≤ 1 000, 0 ≤ M ≤ 10 000, 0 ≤ K < N)
- The next `M` lines: `u v fare`, a ride from stage `u` to stage `v` (u ≠ v) costing `fare` shillings (1 ≤ fare ≤ 10 000). There can be several rides between the same two stages.
- The last line: `S` and `D`, the start and the destination.

**Output**

One line: the cheapest total fare from `S` to `D` with at most `K + 1` rides, or `-1` if there's no way. If `S` is `D`, the answer is `0`.

**Example**

```
Input        Output
4 5 1        350
1 2 100
2 3 100
3 4 100
1 3 250
1 4 500
1 4
```

The cheapest route, 1 → 2 → 3 → 4 for 300, needs two changes. With one change allowed, 1 → 3 → 4 costs 350, which beats the direct ride for 500. With `K = 0`, the answer would be 500.

**Limits:** 2 seconds per test case.

**Go tip**

Plain Dijkstra finds the cheapest route but ignores the number of rides, so it gives 300 here. Two ways that work:

- **Bellman-Ford, stopped early.** Run `K + 1` rounds. In each round, relax every ride using a copy of the costs from the round before, so one round adds at most one ride. That's O(K × M).
- **Dijkstra on (stage, rides used).** Each state is a pair, and you stop expanding a state once it has used `K + 1` rides.

Explain which you'd pick and why before you start typing. In an interview, that conversation earns as many marks as the code.

Judge it from the repo root with `go run ./tools/kata 99`.
