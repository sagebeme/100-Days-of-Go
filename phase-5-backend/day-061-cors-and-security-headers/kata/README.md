# Kata: Sliding Window Limiter

Day 61's limiter is a token bucket. The other common design is a **sliding log**: remember when each IP's allowed requests happened, and allow a new one only if fewer than `k` of them fall in the last `w` seconds. Replay a request log through one and print each verdict.

A request from `ip` at time `t` is allowed if that IP has fewer than `k` **allowed** requests at times `s` with `t - w < s ≤ t`. Denied requests don't count towards the limit. Each IP has its own log.

**Input**

- Line 1: `n w k` (1 ≤ n ≤ 200 000, 1 ≤ w ≤ 10^6, 1 ≤ k ≤ 1 000)
- Next `n` lines: `t ip`, where `t` is a time in seconds (0 ≤ t ≤ 10^9, never decreasing) and `ip` is a string without spaces

**Output**

`n` lines: `ALLOW` or `DENY`.

**Example**

```
Input           Output
7 10 2
1 41.90.1.1     ALLOW
2 41.90.1.1     ALLOW
3 41.90.1.1     DENY
5 41.90.2.2     ALLOW
11 41.90.1.1    ALLOW
12 41.90.1.1    ALLOW
12 41.90.1.1    DENY
```

At time 11 the window is `(1, 11]`, so the request at 1 has dropped out and only the one at 2 is left: allowed. At 12 the window `(2, 12]` holds 11 and 12, which is `k`, so the second request at 12 is denied.

**Limits:** 2 seconds per test case.

**Go tip:** a `map[string][]int` holds each IP's times. Because time never goes backwards, expired times are always at the front, so drop them with `times = times[1:]` until the first one is inside the window. Each time is added once and dropped at most once, so the whole replay is O(n). Store the slice back in the map after changing it.

Judge it from the repo root with `go run ./tools/kata 61`.
