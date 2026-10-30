# Day 92: Log Analyser

## The brief

Otieno is on call for a payments API in Nairobi. Customers say "M-Pesa prompts are slow today", and the only evidence is an nginx access log: 4 GB a day, gzipped, on a server with 1 GB of RAM. He wants a small CLI he can point at a log, or pipe into, that prints the status codes, the busiest paths and the p95 latency in a few seconds, without loading the file into memory, and without falling over on the odd corrupted line.

```
$ logstat -top 3 access.log
lines       12
parsed      9
malformed   3 (first at lines 5, 10, 11)
from        2026-10-24T09:00:01+03:00
to          2026-10-24T09:00:11+03:00
bytes sent  877
1xx         0
2xx         5
3xx         0
4xx         2
5xx         2
p50         50ms
p95         10s
p99         10s
top paths
  4         /stk/push
  3         /balance
  1         /health

$ zcat access.log.gz | logstat -max-5xx 1 -    →  exit status 3 when over 1% of requests were 5xx
```

The log is nginx's `combined` format with `$request_time` (seconds) added at the end:

```
41.90.64.7 - - [24/Oct/2026:09:00:01 +0300] "POST /stk/push HTTP/1.1" 200 312 "-" "okhttp/4.12.0" 0.120
```

Acceptance criteria:

- The file is read as a stream. Memory stays the same whether it's 12 lines or 12 million.
- Malformed lines (junk, impossible dates, status 999, cut-off lines, a 1 MB line) are counted and skipped, and the first few line numbers are reported. Only a read error stops the run.
- Paths are counted without their query string. Ties in the top list are broken by path, so output never changes between runs.
- p50, p95 and p99 use the nearest-rank method and are exact to the millisecond.
- The CLI reads a file or stdin (`-`), has `-top`, `-json` and `-max-5xx` flags, and uses exit codes a script can check.

`starter/testdata/access.log` is the 13-line fixture the tests use, with three bad lines and a blank one.

## What you'll use

- `bufio` for reading: [Day 12](../../phase-1-terminal/day-012-guessing-game/). Flags and exit codes: [Day 22](../../phase-2-modelling/day-022-cli-flags/).
- `text/tabwriter` for the table: [Day 11](../../phase-1-terminal/day-011-receipt-printer/).
- Frequency counts with maps: [Day 28](../../phase-3-algorithms/day-028-duplicate-receipts/). Status classes, as in the Status Board kata: [Day 51](../../phase-5-backend/day-051-status-codes/).
- `time.Parse` with a layout, and table tests over `testdata/`: [Day 19](../../phase-2-modelling/day-019-table-tests/).
- New today: `bufio.Reader.ReadSlice`, which hands back a line without copying and returns `bufio.ErrBufferFull` when a line is longer than the buffer, so one huge line can't eat your memory. And a **histogram** for percentiles: one counter per millisecond, so p95 needs no list of every latency.

```go
var buckets [60_001]int64 // 0 ms to 60 s
buckets[latency/time.Millisecond]++
// p95: walk the buckets until you've passed ceil(0.95 * n) requests
```

## Why it matters at work

Averages hide the pain: if 95 requests take 50 ms and 5 take 10 s, the mean is about 550 ms, which looks fine and is wrong for everyone. That's why dashboards and SLOs use p95 and p99. When the dashboard is down or doesn't have the field you need, the raw log is the source of truth, and a tool that reads it as a stream works on the production box over SSH, where you can't copy 4 GB to your laptop. The engineering habits are the ones reviewers look for in any data pipeline: never trust input, count what you skip, and never let one bad record stop the run.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `ParseLine` follows its doc comment and wraps every error in `ErrMalformed`. The user agent can contain spaces; quoted fields end at the next `"`.
2. `Histogram`: `Add` clamps to 0..`MaxTracked`, and `Percentile` uses the nearest rank, `ceil(p/100 × n)`. Watch floating-point error at the ceiling.
3. `Analyse` reads with a `bufio.Reader` of `MaxLineBytes`, line by line. It skips blank lines, handles `\r\n` and a last line without a newline, counts malformed lines, and skips over-long lines without keeping them.
4. It counts statuses, paths (without `?query`), bytes and the time range, then sorts the top paths by count and then path, and keeps `topN`.
5. `Class` and `Write` as described.
6. `cmd/logstat` reads a file or `-` for stdin. `-json` prints the report as JSON, and `-max-5xx` exits with status 3 when the share of 5xx is higher.
7. Run the tests, then try it:

```bash
go test ./phase-8-portfolio/day-092-log-analyser/starter
go run ./phase-8-portfolio/day-092-log-analyser/starter/cmd/logstat -top 3 \
    phase-8-portfolio/day-092-log-analyser/starter/testdata/access.log
```

Stretch goals:

- Read `.gz` files directly with `compress/gzip`, chosen by the file extension.
- `-since` and `-until` flags, and a per-minute breakdown of requests and 5xx.
- Percentiles per path, for the top paths only, to keep memory bounded.
- Parse with several goroutines: one reader, a pool of parsers, one aggregator ([Day 46](../../phase-4-concurrency/day-046-worker-pools/)). Measure it with a benchmark before you believe it's faster.
- A `-format` flag for other layouts, such as JSON lines from Go's `log/slog`.

## The kata: Busiest Minute

Given every request's time in a day, sorted, find the 60-second window with the most requests. Two pointers over a sorted stream, the same idea as [Day 30](../../phase-3-algorithms/day-030-sliding-window/). Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 92
```

## When you're stuck

- **`bufio.Scanner: token too long`**: `Scanner` gives up on a line over its buffer and stops the whole scan. Use `bufio.Reader.ReadSlice`, and on `bufio.ErrBufferFull` keep reading until the newline, then count the line as malformed.
- **The user agent comes out as `"Mozilla/5.0`**: you split the line on spaces. Read quoted fields from one `"` to the next.
- **`parsing time "31/Feb/2026:09:00:09 +0300": day out of range`**: that's the fixture's bad date doing its job. Count it as malformed.
- **`/balance?account=1234` and `/balance` are counted separately**: cut the target at `?` with `strings.Cut`.
- **p95 is 2.3s, not 10s, on the fixture**: you rounded the rank down. Nine values at 95% is rank 8.55, so take the 9th.
- **The top paths change order between runs**: map iteration is random. Sort by count, then by path.
- **`0.120` becomes 119ms**: floating-point. Round with `math.Round(secs * 1000)` before turning it into a `time.Duration`.
- **The read-error test passes `nil`**: you treated every error like the end of the input. Only `io.EOF` is the end; return anything else.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Give it its own repo with a README: the supported format, a sample report, and how the histogram keeps memory constant.
- Release binaries with GoReleaser so it can be copied to a server, and show a GIF of it reading a large log in a terminal.
- Generate a 1 GB synthetic log in CI, and publish the time and peak memory (`/usr/bin/time -v`) in the README.
- Write "What I'd do next": gzip, time filters, per-path percentiles, concurrent parsing.
- Add GitHub Actions running `go test -race` and a benchmark, with the CI badge.
