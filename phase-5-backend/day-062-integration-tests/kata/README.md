# Kata: Flaky Tests

Integration tests that use real servers, real ports and goroutines can be **flaky**: they pass and fail on the same code. CI reruns failed jobs, so the logs hold several results per commit. Find the flaky tests, so someone can fix them before the team learns to ignore red builds.

A test is flaky on a commit if, on that commit, it has at least one `PASS` and at least one `FAIL`. A test that only ever fails on a commit is broken there, not flaky.

**Input**

- Line 1: `n`, the number of results (1 ≤ n ≤ 200 000)
- Next `n` lines: `commit test result`, where `result` is `PASS` or `FAIL`, and names have no spaces

**Output**

One line per flaky test: its name and the number of commits it was flaky on, separated by a space. Sort by that number, biggest first, then by name in alphabetical (byte) order. If no test is flaky, print `NONE`.

**Example**

```
Input                                  Output
8                                      TestLastSeatRace 2
a1f3 TestBookingJourney PASS
a1f3 TestLastSeatRace FAIL
a1f3 TestLastSeatRace PASS
b7c2 TestLastSeatRace PASS
b7c2 TestHealth FAIL
b7c2 TestHealth FAIL
c9d0 TestLastSeatRace FAIL
c9d0 TestLastSeatRace PASS
```

`TestLastSeatRace` passed and failed on `a1f3` and on `c9d0`. `TestHealth` failed twice on `b7c2` but never passed there, so it's broken, not flaky.

**Limits:** 2 seconds per test case.

**Go tip:** a struct with two string fields can be a map key: `map[run]int` where `type run struct{ commit, test string }`. Store a small bit set per run (`|= 1` for a pass, `|= 2` for a fail), and a run is flaky when its value is 3. Then count per test and sort with `sort.Slice` and a two-level comparison, as on Day 10.

Judge it from the repo root with `go run ./tools/kata 62`.
