# Kata: Replay Window

A payment provider sometimes sends the same webhook twice: a retry after a timeout, or an attacker replaying a captured request. The relay remembers each event ID for `W` seconds. An event is a duplicate if the same ID was accepted less than `W` seconds ago. A duplicate doesn't restart the clock: only accepted events do.

**Input**

- Line 1: `N` and `W`, the number of events and the window in seconds (1 ≤ N ≤ 200 000, 1 ≤ W ≤ 10^9)
- The next `N` lines: `t id`, the time in seconds (0 ≤ t ≤ 10^9) and the event ID (1 to 20 letters, digits and `_`)
- The times never go down, but several events can share a time.

**Output**

`N` lines: `ACCEPT` or `DUPLICATE` for each event, in order.

**Example**

```
Input        Output
7 300        ACCEPT
100 evt_a    ACCEPT
130 evt_b    DUPLICATE
250 evt_a    DUPLICATE
399 evt_a    ACCEPT
400 evt_a    DUPLICATE
500 evt_a    ACCEPT
701 evt_a
```

`evt_a` is accepted at 100, so it's a duplicate until 399. At 400, 300 seconds have passed and it's accepted again, which starts a new window that lasts until 699. At 701 it's accepted once more.

**Limits:** 2 seconds per test case.

**Go tip**

A map from ID to the time it was last accepted is enough. The check is `t-last < w`, strictly less: at exactly `w` seconds, the window is over.

```go
if last, ok := accepted[id]; ok && t-last < w { /* duplicate */ }
```

This is LeetCode's "logger rate limiter". In the real relay the map would grow for ever, so the Deduper also drops old IDs; try adding that with a queue of `(time, id)` pairs, oldest first.

Judge it from the repo root with `go run ./tools/kata 95`.
