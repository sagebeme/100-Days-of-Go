# Kata: Attack Paths

A threat model is a graph: components are nodes and data flows are one-way edges. An attacker starts at anything facing the internet and moves along flows: if they control a component, they can try the next one it sends data to. For each sensitive asset, how many hops is it from the nearest internet-facing component? The closer it is, the fewer defences stand in the way.

**Input**

- Line 1: `n m k`, the number of components, flows and assets to ask about (1 ≤ n ≤ 20 000, 0 ≤ m ≤ 40 000, 1 ≤ k ≤ n)
- Next `n` lines: `name zone`. A component whose zone is exactly `internet` is a starting point.
- Next `m` lines: `from to`, a flow from one component to another
- Next `k` lines: an asset name

Names and zones are 1 to 20 letters, digits or hyphens, and component names are all different. Every flow and asset names a component.

**Output**

For each asset, in input order: `name hops`, the fewest flows from any internet component to it (0 if it is one), or `name SAFE` if none can reach it.

**Example**

```
Input                     Output
5 4 3                     postgres 3
app internet              app 0
api dmz                   backups SAFE
seats internal
postgres internal
backups internal
app api
api seats
seats postgres
backups postgres
postgres
app
backups
```

`app → api → seats → postgres` is 3 hops. `backups` only sends data to `postgres`; nothing flows into it, so it can't be reached.

**Limits:** 2 seconds per test case.

**Go tip:** this is breadth-first search from several starts at once (Day 37). Put every internet component in the queue with distance 0 before you begin, then run one ordinary BFS. Give each name an index with a `map[string]int` and keep the edges as `[][]int`. Use `-1` in the distance slice for "not reached".

Judge it from the repo root with `go run ./tools/kata 82`.
