# Kata: Flaky Finder

CI has run the front-end tests many times this week. Some tests pass every time, some fail every time, and some do both: they're **flaky**, usually because they depend on timing, the clock, or the order tests run in. Flaky tests are worse than no tests, because people learn to ignore red. Find them.

**Input**

- Line 1: `n`, the number of results (1 ≤ n ≤ 10 000)
- Next `n` lines: `name result`, where `name` is one word and `result` is `PASS` or `FAIL`

**Output**

Go through the test names in sorted order (byte order, as `sort.Strings` gives, so capital letters come before small ones):

- a test that both passed and failed: `flaky NAME P/T`, where `P` is its passes and `T` its runs
- a test that never passed: `broken NAME`
- a test that always passed: print nothing for it

Finish with `stable K`, the number of tests that always passed.

**Example**

```
Input                  Output
6                      broken login_errors
trips_loads PASS       flaky trips_loads 2/3
login_errors FAIL      stable 1
trips_loads FAIL
seats_pick PASS
login_errors FAIL
trips_loads PASS
```

**Limits:** 2 seconds per test case.

**Go tip:** two maps, `runs := map[string]int{}` and `passes := map[string]int{}`. Map order is random in Go, so collect the keys into a slice and `sort.Strings` it before printing.

Judge it from the repo root with `go run ./tools/kata 73`.
