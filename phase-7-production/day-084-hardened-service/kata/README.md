# Kata: Log Aggregation

It's Monday morning after the hardened Twende API's first weekend. The structured logs are in, one line per request. The SACCO manager wants a one-screen summary: per route, how many requests, how many failed, and how slow the slow ones were; and when, to the minute, things were worst.

**Input**

- Line 1: `n`, the number of log lines (1 ≤ n ≤ 100 000)
- Then `n` lines: `ts route status ms`. `ts` is seconds since the logs began (0 ≤ ts ≤ 10^9, not necessarily in order), `route` is 1 to 30 characters with no spaces, `status` is an HTTP status code (100 to 599) and `ms` is the latency in milliseconds (0 ≤ ms ≤ 10^6).

**Output**

- One line per route, sorted by route name (byte order): `route requests errors p95`. `errors` counts status 500 and above (a 4xx is the client's mistake, not an outage). `p95` is the nearest-rank 95th percentile of that route's latencies (Day 77): sort them and take the one at position `⌈95 × count / 100⌉`, counting from 1.
- Then `ALL requests errors p95` over every line.
- Then `worst minute M with E errors`, where minute `M` is `ts / 60` rounded down and `E` is the most 5xx errors in any one minute; on a tie, the earliest minute. If there are no errors at all, print `worst minute none`.

**Example**

```
Input                         Output
6                             /bookings 3 2 900
0 /trips 200 40               /trips 3 0 70
61 /bookings 503 900          ALL 6 2 900
59 /trips 200 70              worst minute 1 with 2 errors
75 /bookings 500 120
62 /bookings 201 80
130 /trips 404 15
```

`/bookings` failed at 61 s and 75 s, both in minute 1. The 404 on `/trips` isn't an error.

**Limits:** 2 seconds per test case.

**Go tip:** keep a `map[string][]int` of latencies per route and a `map[int]int` of errors per minute. Collect the route names into a slice and `slices.Sort` it before printing, since map order is random. `slices.Sort` each latency slice once and index it: `(95*len(l) + 99) / 100` is the rank in whole numbers. Read with a `bufio.Reader` and write with a `bufio.Writer`.

Judge it from the repo root with `go run ./tools/kata 84`.
