# Day 40: Mock HackerRank Test

## The brief

This is the real thing, as close as the course can make it. A Nairobi fintech has sent you an online assessment: seven problems, 90 minutes, judged automatically. Everything from Days 27 to 39 is fair game, and nobody tells you which technique each problem needs. Spotting that is half the test.

Set a timer for **90 minutes** before you open `starter/mock.go`. Don't look at `solution/`. When the timer goes, stop, run everything, and score yourself.

| # | Problem | Where | Points |
|---|---|---|---|
| 1 | Missing Receipt | `starter/mock.go` | 10 |
| 2 | First Unique | `starter/mock.go` | 10 |
| 3 | Best Trade | `starter/mock.go` | 15 |
| 4 | Merge Bookings | `starter/mock.go` | 15 |
| 5 | Count Shambas | `starter/mock.go` | 15 |
| 6 | Longest Rising | `starter/mock.go` | 15 |
| 7 | Jam Session | `kata/` | 20 |

```
MissingReceipt([]int{3, 1, 4})                                 →  2
FirstUnique("mamboleo")                                        →  1
BestTrade([]int{20, 18, 25, 17, 22, 30, 21})                   →  13
MergeBookings([]Interval{{8, 10}, {1, 3}, {2, 6}, {15, 18}})   →  [{1 6} {8 10} {15 18}]
CountShambas([]string{"##..", "#..#", "..##"})                 →  2
LongestRising([]int{10, 9, 2, 5, 3, 7, 101, 18})               →  4
```

Problem 7 is a stdin/stdout program, like every kata: the fewest minutes across a city grid in Friday traffic.

## What you'll use

All of Phase 3. Keep [reference/big-o.md](../../reference/big-o.md) open: a real test lets you have notes far less often than you'd like, so practise using only that one page.

### How to take a timed test

1. **Read every problem first (5 minutes).** Note each one's constraints and your first idea. Some later problems are easier than earlier ones. Knowing all of them lets you spend your time where the points are.
2. **Read the constraints before you design.** Turn the largest `n` into a budget with the 10^8 rule: n = 10^6 needs O(n) or O(n log n), n = 2 × 10^5 rules out O(n²), a grid of 1000 × 1000 is 10^6 cells. The constraints often tell you the technique.
3. **Start with what you can finish.** Two problems solved fully beat four half-done. Bank the easy points in the first 20 minutes.
4. **Write the brute force when you're stuck.** It earns partial credit on real platforms, where tests are weighted and the small ones often pass with an O(n²) answer. It also checks your fast version.
5. **List the edge cases before you run the tests:** empty input, one element, all equal, already sorted, reverse sorted, negatives, the largest values (overflow?), no answer at all (-1, 0 or empty?). Most hidden test cases are edge cases.
6. **Use the samples, then make your own.** Passing the examples isn't passing the test.
7. **Watch the clock.** If a problem has taken 20 minutes and you have no plan, move on and come back.
8. **Fast I/O for stdin problems.** `bufio` in, `bufio` out, `defer out.Flush()`. A correct answer with `fmt.Scan` on 10^6 numbers can still time out.

## Steps

There's no walkthrough today. These are the rules of the test.

1. Start a 90-minute timer.
2. Solve problems 1 to 6 in `starter/mock.go`, in any order. Each doc comment gives the problem, its constraints and an example. Don't change the signatures.
3. Solve problem 7 in `kata/main.go`. The statement is in `kata/README.md`.
4. Every big test in `starter/mock_test.go` is sized so that the slow answer fails or takes far too long. A solution that only passes the small tables isn't done.
5. Run one problem's tests with `-run`, or all of them, from the repo root:

```bash
go test -run Problem3 ./phase-3-algorithms/day-040-mock-hackerrank-test/starter
go test ./phase-3-algorithms/day-040-mock-hackerrank-test/starter
go run ./tools/kata 40
```

6. **Scoring:** a problem's points count only if all its tests pass. 70 or more is a pass. 85 or more and you're ready for real assessments. Whatever you score, retake the test in a week without looking at your first attempt.

## The kata: Jam Session

Problem 7, the hardest one: the fewest minutes across a city grid where every block has its own traffic delay. BFS isn't enough because the costs differ, and right-and-down DP isn't enough because the best road winds. It's Dijkstra with `container/heap`. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 40
```

## When you're stuck

These hints name the technique. In a real test, nobody would tell you, so try each problem for at least ten minutes first.

- **Problem 1:** the numbers 1 to n+1 add up to a formula ([Day 27](../day-027-how-fast-is-fast/)). Subtract what you have. Sorting is O(n log n) and not needed.
- **Problem 2:** two passes. Count every byte first ([Day 28](../day-028-duplicate-receipts/): an array of 256 counts works as a map for bytes), then find the first count of 1.
- **Problem 3:** one pass. Keep the cheapest price seen so far. Each day, the best you could do selling today is today's price minus that. A nested loop is O(n²) and fails the million-day test.
- **Problem 4:** sort by start ([Day 31](../day-031-sorting-showdown/)), clone first so the input doesn't change. Then the next booking either touches the last merged one (`Start <= End`) or starts a new one.
- **Problem 5:** each unvisited `#` starts a new shamba. Flood fill from it ([Day 37](../day-037-graphs/)) and mark everything you reach. Use an explicit stack or a queue, not recursion: the snake test goes 500 000 cells deep.
- **Problem 6:** the O(n²) DP is the one in textbooks, and too slow. Keep `tails`, where `tails[k]` is the smallest last value of any rising run of length k+1. It stays sorted, so binary search ([Day 32](../day-032-binary-search/): `sort.SearchInts`) finds where each value goes. Equal values replace, they don't extend.
- **Problem 7, wrong answer on a test with a winding road:** you only moved right and down. Traffic can make going back up faster.
- **Problem 7, too slow:** check that you skip stale heap entries, and that you mark cells as numbers `r*C + c` in a slice, not as keys in a map.
- **Still stuck when the timer goes?** That's information, not failure: note which days to revisit. Then open `solution/mock.go` and `kata/solution/main.go`, read each answer once, close them, and write your own version from memory.
