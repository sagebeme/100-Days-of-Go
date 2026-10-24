# Kata: Busiest Windows

A bank branch in Thika records the length of its queue every minute as a single digit, `0` to `9`. The manager looks at every window of `k` consecutive minutes and writes down the longest queue in that window. Print the sum of everything she writes down.

**Input**

- Line 1: `n` and `k` (1 ≤ k ≤ n ≤ 2 × 10^5)
- Line 2: `n` digits with no spaces, the queue length each minute

**Output**

One line: the sum of the maximum of every window of `k` consecutive minutes. There are `n - k + 1` windows.

**Example**

```
Input           Output
7 3             31
3141592
```

The windows are `314`, `141`, `415`, `159`, `592`. Their maxima are 4, 4, 5, 9, 9, which add up to 31.

**Limits:** 2 seconds per test case.

**Read the constraints.** Scanning each window is O(n·k). With n = 2 × 10^5 and k = 10^5 that's 10^10 steps. The sliding window maximum with a deque is O(n): keep indexes whose values fall from front to back. Before pushing minute `i`, pop from the back every index whose value is no bigger, since it can never be a maximum again. Pop the front once it slides out of the window. The front is then the window's maximum.

**Go tip**

`log[i] - '0'` turns a digit byte into its value, as a `byte`. Wrap it in `int(...)` before adding it to an `int` total. You can use today's `Deque[int]`, or a plain slice with a head and a tail index, since every index is pushed once.

Judge it from the repo root with `go run ./tools/kata 34`.
