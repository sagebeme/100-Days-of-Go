# Kata: Impatient Queue

One teller at a SACCO banking hall serves customers in the order they arrive. Customer `i` arrives at minute `a_i`, needs `s_i` minutes at the counter, and will wait at most `p_i` minutes for their turn. If the teller can't start on them within `p_i` minutes of arriving, they give up and leave, and the teller never spends time on them. Waiting exactly `p_i` minutes is fine.

It's a `select` with a timeout, in numbers: each customer waits for either the teller or their own deadline, whichever comes first.

Print how many customers were served, how many left, and the minute the teller finishes the last customer they served (0 if they served nobody).

**Input**

- Line 1: `n`, the number of customers (1 ≤ n ≤ 200 000)
- The next `n` lines: `a_i s_i p_i` (0 ≤ a_i ≤ 10^9, 1 ≤ s_i ≤ 10^4, 0 ≤ p_i ≤ 10^9), with the `a_i` in non-decreasing order

**Output**

- Line 1: the number served and the number who left, separated by a space
- Line 2: the minute the teller finishes their last customer

**Example**

```
Input           Output
4               2 2
0 5 10          9
1 3 2
2 4 10
3 2 1
```

The first customer is served at once and the teller is busy until 5. The second would start at 5 after waiting 4 minutes, more than their 2, so they leave. The third waits 3 minutes, which is fine, and is done at 9. The fourth would wait 6 minutes, more than 1: they leave.

**Limits:** 2 seconds per test case.

**Go tip**

Keep one variable, `free`, the minute the teller is next free. Customer `i` would start at `max(a_i, free)`. Only a served customer moves `free` forward. Go 1.21 added the built-in `max` and `min`, so you don't have to write them.

Judge it from the repo root with `go run ./tools/kata 43`.
