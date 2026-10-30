# Kata: Busiest Minute

On-call wants to know when the traffic spike hit. Given the time of every request in one day's log, find the 60-second window with the most requests. It's a sliding window with two pointers, the pattern from Day 30, over a stream you've already sorted.

A window starts at some request's time `t` and covers `t` to `t + 59` seconds, inclusive.

**Input**

- Line 1: `n` (1 ≤ n ≤ 200 000)
- The next `n` lines: a time `HH:MM:SS` (00:00:00 to 23:59:59), in order, earliest first. Times can repeat.

**Output**

- One line: the most requests in any window, a space, and the start of that window as `HH:MM:SS`. If several windows tie, print the earliest.

**Example**

```
Input       Output
7           4 09:00:30
09:00:01
09:00:02
09:00:30
09:01:01
09:01:02
09:01:03
09:02:00
```

From 09:00:01 the window runs to 09:01:00 and holds 3 requests. From 09:00:30 it runs to 09:01:29 and holds 4: 09:00:30, 09:01:01, 09:01:02 and 09:01:03. No window holds more.

**Limits:** 2 seconds per test case. Checking every window from scratch is O(n²), which is 4 × 10^10 steps at the limit. Move the right pointer forward only.

**Go tip**

Turn each time into seconds once, with `fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)`, and print the answer back with `%02d`:

```go
fmt.Printf("%d %02d:%02d:%02d\n", best, t/3600, t/60%60, t%60)
```

Judge it from the repo root with `go run ./tools/kata 92`.
