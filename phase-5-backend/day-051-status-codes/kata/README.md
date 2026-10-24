# Kata: Status Board

It's Monday morning and the on-call engineer wants one screen: how many requests in each status class, how bad the server errors were, and which path broke most often. Build it from the access log.

**Input**

- Line 1: `n`, the number of log lines (1 ≤ n ≤ 5 000)
- The next `n` lines: `method path status ms`, separated by single spaces. `status` is from 200 to 599. `ms` is the time taken, which you can ignore.

**Output**

Six lines:

1. to 4. `2xx c`, `3xx c`, `4xx c`, `5xx c`: the number of lines in each class, in that order, even when a count is 0
5. `5xx rate P%`: the share of lines that were 5xx, as a percentage **cut down (not rounded)** to 2 decimal places, always with 2 digits after the point
6. `worst path k`: the path with the most 5xx answers and how many it had. On a tie, the path that comes first in alphabetical (byte) order. If there were no 5xx at all, print `worst none`

**Example**

```
Input                              Output
7                                  2xx 2
GET /trips/1 200 12                3xx 0
POST /trips/1/bookings 201 40      4xx 2
POST /trips/1/bookings 409 8       5xx 3
GET /bookings/7 404 3              5xx rate 42.85%
POST /trips/2/bookings 500 1203    worst /trips/2 2
GET /trips/2 503 30
GET /trips/2 503 31
```

3 out of 7 is 42.857...%, cut down to `42.85%`.

**Limits:** 2 seconds per test case.

**Go tip**

Floats can turn 42.85 into 42.849999. Stay in whole numbers: `v := c5 * 10000 / n` is the percentage times 100, cut down, and `%02d` pads the decimals:

```go
v := 3 * 10000 / 7                       // 4285
fmt.Printf("%d.%02d%%\n", v/100, v%100)  // 42.85%
```

`%%` prints a literal percent sign. The class of a status is `status / 100`.

Judge it from the repo root with `go run ./tools/kata 51`.
