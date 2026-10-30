# Kata: Within K Changes

A classic interview graph problem with a twist on Dijkstra. Stages are numbered `0` to `n-1`. Each matatu goes one way from stage `u` to stage `v` for a fare. Find the cheapest way from `s` to `d` using **at most `k` changes**, which means at most `k + 1` matatus.

Plain Dijkstra on fares doesn't work: the cheapest way to reach a middle stage might use up all your changes. Relax every edge `k + 1` times instead (Bellman-Ford with a limit), copying the distances before each round so one round adds at most one matatu.

**Input**

- Line 1: `n m k s d` (1 ≤ n ≤ 100, 0 ≤ m ≤ 10 000, 0 ≤ k < n, 0 ≤ s, d < n)
- The next `m` lines: `u v w`, a matatu from `u` to `v` costing `w` shillings (1 ≤ w ≤ 10 000, u ≠ v)

**Output**

- The cheapest total fare, or `-1` if `d` can't be reached within the limit. From a stage to itself costs `0`.

**Example**

```
Input       Output
3 3 1 0 2   200
0 1 100
1 2 100
0 2 500
```

With one change allowed, `0 → 1 → 2` costs Ksh 200. With `k = 0` the only choice would be the direct Ksh 500 matatu.

**Limits:** 2 seconds per test case. `(k + 1) × m` is at most a million relaxations.

**Go tip**

Copy the slice at the start of each round with `slices.Clone`, read from the old one and write to the new one:

```go
next := slices.Clone(dist)
for _, e := range edges {
	if dist[e.u] != inf && dist[e.u]+e.w < next[e.v] {
		next[e.v] = dist[e.u] + e.w
	}
}
dist = next
```

Judge it from the repo root with `go run ./tools/kata 91`.
