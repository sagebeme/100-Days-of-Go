# Kata: Kata Leaderboard

The bot keeps a leaderboard for the server. Each solve is reported once per user and day; a user who solves the same day again gets nothing more. `TOP k` asks for the best `k` users so far.

Rank users by:

1. more katas solved first,
2. then fewer hints used in total,
3. then name, in alphabetical (byte) order.

**Input**

- Line 1: `Q`, the number of lines that follow (1 ≤ Q ≤ 200 000)
- The next `Q` lines, one of:
  - `SOLVE user day hints`: `user` solved kata `day` after `hints` hints (1 ≤ day ≤ 100, 0 ≤ hints ≤ 10). Only the first solve of each `(user, day)` counts.
  - `TOP k`: print the leaderboard's first `k` users (1 ≤ k ≤ 10)
- Names are 1 to 20 lowercase letters and digits. There are at most 1 000 users and at most 1 000 `TOP` lines.

**Output**

For each `TOP k`, one line: up to `k` entries `user:solved:hints`, separated by single spaces, best first. If nobody has solved anything yet, print `-`.

**Example**

```
Input               Output
10                  -
TOP 3               amina:2:1 juma:1:0
SOLVE amina 93 1    amina:2:1 juma:2:2
SOLVE juma 93 0     amina:2:1
SOLVE amina 94 0
TOP 2
SOLVE juma 94 2
SOLVE juma 94 0
TOP 5
SOLVE wanjiru 95 1
TOP 1
```

Juma's second solve of day 94 doesn't count, so he stays on 2 hints. With two solves each, Amina's 1 hint beats Juma's 2.

**Limits:** 2 seconds per test case.

**Go tip**

Keep a map from `(user, day)` to "already counted", using a struct as the map key. With at most 1 000 users and 1 000 `TOP` lines, sorting a copy of the board for each `TOP` is fast enough. `cmp.Or` (Go 1.22) returns its first non-zero argument, which makes a three-level sort one line each:

```go
slices.SortFunc(ranked, func(a, b standing) int {
	return cmp.Or(cmp.Compare(b.solved, a.solved), cmp.Compare(a.hints, b.hints), strings.Compare(a.user, b.user))
})
```

For bigger limits, keep the board sorted as you go, or use a heap for the top `k`.

Judge it from the repo root with `go run ./tools/kata 96`.
