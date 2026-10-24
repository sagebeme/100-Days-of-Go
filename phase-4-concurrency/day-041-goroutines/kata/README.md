# Kata: Batch Sender

An SMS gateway sends messages in batches of `k`. It starts the first `k` messages together, waits until the slowest of them is done, then starts the next `k`, and so on. The last batch may be smaller. Message `i` takes `t_i` milliseconds.

Print how long everything takes when the messages go one at a time, and how long it takes in batches.

**Input**

- Line 1: `n` and `k`, the number of messages and the batch size (1 ≤ n ≤ 200 000, 1 ≤ k ≤ 200 000)
- Line 2: `n` durations `t_1 ... t_n` in milliseconds (1 ≤ t_i ≤ 1 000 000), in sending order

**Output**

- Line 1: the total time one at a time: the sum of all the durations
- Line 2: the total time in batches: for each batch, the duration of its slowest message, added up

**Example**

```
Input           Output
5 2             16
3 1 4 1 7       14
```

One at a time: 3 + 1 + 4 + 1 + 7 = 16. In batches of 2 the gateway sends `[3 1]`, then `[4 1]`, then `[7]`. Each batch lasts as long as its slowest message: 3 + 4 + 7 = 14.

**Limits:** 2 seconds per test case. Totals can reach 2 × 10^11, which needs Go's 64-bit `int` (it is 64 bits on every computer you'll use).

**Go tip**

This is `wg.Wait()` in numbers: a batch is only done when its slowest goroutine is. Keep a running maximum, and add it to the total each time you reach the end of a batch:

```go
if i%k == k-1 || i == n-1 { // the last message of a batch
	batched += slowest
	slowest = 0
}
```

Judge it from the repo root with `go run ./tools/kata 41`.
