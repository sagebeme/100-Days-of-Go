# Kata: Token Buckets

Simulate the proxy's limiter without a clock. Each client has a bucket that holds up to `b` tokens and starts full. Tokens drip back in at `r` per second, continuously, never above `b`. A request takes one token if there is at least one whole token; otherwise it's denied and takes nothing.

Do it in whole numbers: count in thousandths of a token. Then `r` tokens per second is exactly `r` thousandths per millisecond, and nothing is ever rounded.

**Input**

- Line 1: `b`, `r` and `n` (1 ≤ b ≤ 1 000, 0 ≤ r ≤ 1 000, 1 ≤ n ≤ 200 000)
- The next `n` lines: `t client`, a time in milliseconds (0 ≤ t ≤ 10^12, never decreasing) and a client name with no spaces

**Output**

- One line per request: `ALLOW` or `DENY`
- A last line: how many were allowed

**Example**

```
Input              Output
2 1 6              ALLOW
0 41.90.64.7       ALLOW
0 41.90.64.7       DENY
0 41.90.64.7       DENY
500 41.90.64.7     ALLOW
1000 41.90.64.7    ALLOW
1000 196.201.214.9 4
```

The first client spends both tokens at time 0 and is denied the third. By 500 ms it has half a token: still denied. By 1000 ms it has one whole token. The second client's bucket starts full.

**Limits:** 2 seconds per test case. `r * (t - last)` can reach 10^15, which fits in an `int` but not an `int32`. Cap at `b * 1000` after adding.

**Go tip**

Keep each client's state in a small struct in a map, and only update it when that client shows up:

```go
type bucket struct{ milli, last int }
b, ok := buckets[client]
if !ok {
	b = bucket{milli: cap, last: t}
}
b.milli = min(cap, b.milli+r*(t-b.last))
```

Judge it from the repo root with `go run ./tools/kata 90`.
