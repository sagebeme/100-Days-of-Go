# Kata: Last Write Wins

The seat map arrives as a snapshot, then a stream of updates. On a phone hopping between towers, updates can arrive late and out of order. Each update carries the version number the server gave it. For each seat, keep the update with the highest version, so a late, older update can't undo a newer one.

The snapshot has version `V`. Every seat starts at version `V`. An update `version seat taken` changes that seat only if `version` is greater than the version that seat has now; then the seat takes that version. Otherwise the update is ignored.

**Input**

- Line 1: `S V m`: seats (1 ≤ S ≤ 100), snapshot version (0 ≤ V ≤ 10^9), number of updates (0 ≤ m ≤ 15 000)
- Line 2: the snapshot, `S` characters, `1` for taken and `0` for free (seat 1 first)
- Next `m` lines: `version seat taken`, in the order they arrived (1 ≤ seat ≤ S, `taken` is 0 or 1). No two updates have the same version.

**Output**

- Line 1: the final map, in the same form as the snapshot
- Line 2: `ignored K`, the number of updates that changed nothing because they were too old

**Example**

```
Input      Output
5 7 4      11100
01000      ignored 2
9 3 1
8 3 0
10 1 1
6 5 1
```

Version 9 takes seat 3. Version 8 arrives late for seat 3, so it's ignored. Version 10 takes seat 1. Version 6 is older than the snapshot: ignored.

**Limits:** 2 seconds per test case.

**Go tip:** keep the map as a `[]byte` from the snapshot string and a `[]int` of versions, one per seat. `seats[i] = '0' + byte(taken)` turns 0 or 1 into the character.

Judge it from the repo root with `go run ./tools/kata 71`.
