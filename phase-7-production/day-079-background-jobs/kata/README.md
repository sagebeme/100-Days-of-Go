# Kata: Retry Timeline

Twende's SMS worker sends one message at a time, and each attempt takes exactly 1 second. The SMS gateway is flaky: job number `i` fails its first `f` attempts and then works. After a job's `k`-th failure, which ends at time `t`, it waits and becomes ready again at `t + min(base × 2^(k-1), cap)`. A job that fails `m` times is dead and never runs again. No jitter here, so the answer is exact.

Whenever the worker is free, it starts the ready job with the smallest ready time; on a tie, the one that came first in the input. If nothing is ready, it waits for the next job to become ready. A job is ready first at its arrival time.

Print when each job finished for good.

**Input**

- Line 1: `n m base cap` (1 ≤ n ≤ 20 000, 1 ≤ m ≤ 30, 1 ≤ base ≤ cap ≤ 1 000 000 000)
- Then `n` lines: `id arrive f`, a job's name (1 to 20 letters, digits, hyphens or colons, all different), its arrival time in seconds (0 ≤ arrive ≤ 1 000 000 000) and how many attempts fail (0 ≤ f ≤ 100)

**Output**

`n` lines in input order: `id DONE t` with the time its successful attempt ended, or `id DEAD t` with the time its last failed attempt ended.

**Example**

```
Input             Output
3 3 2 10          a DONE 4
a 0 1             b DEAD 10
b 0 5             c DONE 3
c 2 0
```

At 0, `a` and `b` are both ready and `a` came first. It fails from 0 to 1 and is ready again at 1 + 2 = 3. `b` fails from 1 to 2, ready at 4. `c` arrived at 2 and works from 2 to 3. `a` works from 3 to 4. `b` fails a second time from 4 to 5 and waits 4, so it's ready at 9; the worker is idle until then. At 9 to 10 it fails a third time, and 3 failures is `m`, so it's dead at 10.

**Limits:** 2 seconds per test case. Times can pass 2^32; use `int`.

**Go tip:** keep the ready times in a min-heap with `container/heap`. Your type needs `Len`, `Less`, `Swap`, `Push` and `Pop`; make `Less` compare the ready time first and the input index second. Pop the smallest, set the clock to `max(clock, ready)`, run one attempt, and push the job back if it failed and has attempts left.

Judge it from the repo root with `go run ./tools/kata 79`.
