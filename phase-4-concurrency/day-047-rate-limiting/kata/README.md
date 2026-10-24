# Kata: Gateway Bucket

An SMS gateway guards itself with a token bucket. The bucket holds at most `b` tokens and is full at time 0. Tokens drip in steadily, one every `r` milliseconds, but a full bucket can't hold more, so drips into a full bucket are lost. A message that arrives when the bucket has at least one whole token is sent and uses up one token. Otherwise it is rejected, and uses nothing.

Given when each message arrives, print how many were sent, how many were rejected, and the time of the last one sent.

**Input**

- Line 1: `n`, `b` and `r` (1 ≤ n ≤ 200 000, 1 ≤ b ≤ 10^9, 1 ≤ r ≤ 10^9)
- Line 2: `n` arrival times in milliseconds, in non-decreasing order (0 ≤ t_i ≤ 10^12)

**Output**

- Line 1: the number sent and the number rejected
- Line 2: the arrival time of the last message sent (the first message is always sent, so there is one)

**Example**

```
Input                   Output
6 2 1000                4 2
0 0 0 500 1000 2500     2500
```

At 0 the bucket has 2 tokens: the first two messages are sent and the third is rejected. At 500 there's only half a token: rejected. At 1000 one whole token has dripped in: sent. By 2500 another 1.5 tokens have dripped in: sent.

**Limits:** 2 seconds per test case.

**Go tip**

Avoid fractions. Keep whole `tokens` and `last`, the time the last whole token dripped in (or the time you last found the bucket full). At each arrival `t`, add `(t - last) / r` tokens and move `last` forward by that many steps of `r`, which keeps the leftover part of a step. If that fills the bucket, cap it at `b` and set `last = t`, because a full bucket doesn't save up. It's the same logic as today's `TokenBucket.Allow`.

Judge it from the repo root with `go run ./tools/kata 47`.
