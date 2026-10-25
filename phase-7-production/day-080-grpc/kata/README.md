# Kata: Top Routes

Twende's operations screen streams bookings in as they happen, and the SACCO manager keeps asking "which routes are busiest right now?". Keep a running total of seats booked per route and answer each `top` query with the `k` busiest routes.

Sorting every route for every query is too slow when there are thousands of routes. Keep a **min-heap of size k** instead: push each route, and once the heap holds more than `k`, pop the smallest. What's left is the top `k`. That's O(R log k) per query instead of O(R log R).

**Input**

- Line 1: `n`, the number of lines that follow (1 ≤ n ≤ 20 000)
- Then `n` lines, each one of:
  - `book <route> <seats>`: add `seats` (1 ≤ seats ≤ 100) to the route's total
  - `top <k>`: ask for the `k` busiest routes (1 ≤ k ≤ 10)

Route names are 1 to 20 letters, digits or hyphens. There are at most 3 000 different routes and at most 100 `top` lines.

**Output**

For each `top`, one line: the routes, busiest first, separated by single spaces. Routes with equal totals go in alphabetical order (byte order, as Go compares strings). If fewer than `k` routes have bookings, print all of them. If none have, print `-`.

**Example**

```
Input           Output
7               -
top 3           46K 111
book 46K 4      111 46K 125
book 111 2
top 2
book 111 2
book 125 1
top 5
```

After the first two bookings 46K has 4 and 111 has 2. Then 111 also reaches 4; on a tie, `111` comes before `46K` because `1` is before `4`.

**Limits:** 2 seconds per test case.

**Go tip:** `container/heap` needs a type with `Len`, `Less`, `Swap`, `Push` and `Pop`. For a min-heap that keeps the best `k`, `Less(i, j)` must say "i is worse than j": a smaller total, or an equal total and a name that sorts later. After the loop the heap holds the answer in reverse, so pop everything into a slice and print it backwards.

Judge it from the repo root with `go run ./tools/kata 80`.
