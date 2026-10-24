# Day 77: System Design Basics

## The brief

Twende (the made-up matatu SACCO app from Phase 5) is growing. The SACCO wants to sell seats on every route out of town, and the board asks one question: "Will it fall over at 5 pm on a Friday?" Nobody knows yet. Today you answer with numbers, not hope: you draw the system, then turn the back-of-the-envelope maths into functions with tests, so the next estimate is one function call away.

```
AverageRPS(864_000, 10)                 →  100        requests a second, spread over the day
PeakRPS(864_000, 10, 5)                 →  500        Friday evening is 5× the average
Concurrency(500, 200*time.Millisecond)  →  100        requests in flight at once (Little's law)
ServersNeeded(500, 200, 0.3)            →  4          servers, keeping 30% spare
StorageBytes(50_000, 500, 365, 3)       →  27_375_000_000   bytes of bookings a year, 3 copies
BandwidthMbps(500, 2_000)               →  8          megabits a second out
Summarise(latencies)                    →  {P50: 20ms, P95: 20ms, P99: 2s}
```

## What you'll use

**Drawing the system.** Before any maths, draw boxes and arrows. Every box is something that can fail or slow down; every arrow is a network call that can time out. Twende today:

```
 phones ──HTTPS──▶ load balancer ──▶ api-1 ┐
                                   ──▶ api-2 ┼──writes──▶ Postgres primary
                                   ──▶ api-3 ┘               │ replication
                                        └────reads──▶ read replica
```

- **Load balancer.** One address for the phones; it spreads requests across the API servers and stops sending to one that fails its health check (Day 81).
- **Stateless services.** An API server keeps nothing between requests: no sessions in memory, no files on its own disk. Sessions live in a signed token or a shared store (Day 58). Then any server can answer any request, and you add capacity by adding servers.
- **Read replicas.** Most traffic is "list trips" and "show seats", which only read. A replica is a copy of the database that follows the primary. Reads go there; writes (bookings) go to the primary. A replica can lag behind by a moment, so "show my booking" straight after booking should read the primary.
- **The bottleneck.** You can add API servers easily. The single primary database is the box you can't just copy, so estimate it first.

**Back-of-the-envelope numbers.**

- A day has 86 400 seconds. Average requests per second = daily users × requests each ÷ 86 400.
- Traffic is never even. The **peak factor** says how much busier the rush hour is than the average. 3 to 10 is common; measure yours.
- **Little's law**: requests in flight = arrival rate × time each takes. 500 a second that each take 0.2 s means about 100 at once. That's your connection pool and goroutine count.
- **Headroom**: run servers at about 70% of what they can do, so a spike or a dead server doesn't topple the rest.
- **Percentiles, not averages.** If 98 requests take 20 ms and two take 2 s, the average is about 60 ms and looks fine. p99 (the time 99% of requests are at or under) is 2 s, and those passengers are staring at a spinner. The **nearest-rank** method: sort, then take the sample at position ⌈p/100 × n⌉, counting from 1.

```go
sorted := slices.Clone(samples) // a copy, so the caller's slice keeps its order
slices.Sort(sorted)
rank := int(math.Ceil(0.95 * float64(len(sorted)))) // p95
p95 := sorted[rank-1]
```

`math.Ceil` rounds up. `d.Seconds()` turns a `time.Duration` into a `float64` number of seconds. `int64(x)` before multiplying keeps big storage figures from overflowing on a 32-bit build.

## Why it matters at work

Design reviews start with "how many, how big, how fast?". An engineer who can say "500 requests a second at peak, 100 in flight, 4 servers, and the primary database is the risk" in two minutes gets the review back on track, and one who can't gets asked to come back with numbers. Keeping the estimate in code means it gets updated when the inputs change instead of living in a stale slide. The same habits carry into on-call: alerts and SLOs (Day 81) are written on p95 and p99, because averages hide the passengers having a bad time.

## Steps

1. Open `starter/capacity.go`. Read the doc comments; each one has a worked example.
2. Write `AverageRPS` and `PeakRPS`. Convert to `float64` before dividing, or integer division throws the fraction away. `PeakRPS` should call `AverageRPS`.
3. Write `StorageBytes`. Convert each input to `int64` first, then multiply.
4. Write `BandwidthMbps`: bytes a second × 8 bits ÷ 1 000 000.
5. Write `Concurrency` with Little's law and `latency.Seconds()`.
6. Write `ServersNeeded`. Usable capacity is `perServer × (1 - headroom)`. Return 0 if that's not positive. Round up with `math.Ceil` and never return less than 1.
7. Write `Percentile` with the nearest-rank method. Return `ErrNoSamples` for no samples, clamp the rank to `1..n`, and sort a copy.
8. Write `Summarise`. Sort one copy and read p50, p95 and p99 from it; a small helper that takes a sorted slice keeps it to a few lines.
9. Run the tests from the repo root:

```bash
go test ./phase-7-production/day-077-system-design-basics/starter
```

Then use it: change the inputs to Twende's real guesses (how many passengers, how many taps each) and see which number scares you.

## The kata: Hash Ring

Replay adds, removes and lookups on a consistent hashing ring, the trick caches and sharded databases use so that adding a server moves only a few keys. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 77
```

## When you're stuck

- **`AverageRPS(200000, 6) = 13, want 13.888889`**: you divided ints. Convert with `float64(dailyUsers)` before the division.
- **`StorageBytes(...)` is negative or wrong for the big case**: the multiplication overflowed before you converted. Convert every input to `int64`, then multiply.
- **`ServersNeeded(500, 200, 0.3) = 3, want 4`**: you rounded down with `int(...)`. Use `math.Ceil` first.
- **`ServersNeeded(0, 1000, 0.3) = 0, want 1`**: a service with no traffic still runs on one server. Use `max(n, 1)`.
- **`Percentile(..., 76) = 30ms, want 40ms`**: you rounded the rank down or to nearest. Nearest-rank always rounds up.
- **`index out of range [-1]`**: a tiny `p` gave rank 0. Clamp it to at least 1.
- **`samples became [10ms 20ms 30ms 40ms]`**: you sorted the caller's slice. Sort `slices.Clone(samples)` instead.
- **Still stuck?** Open `solution/capacity.go`, read it once, close it, then write your own version from memory.
