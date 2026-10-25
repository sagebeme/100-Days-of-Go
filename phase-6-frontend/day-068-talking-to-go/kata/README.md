# Kata: Latest Answer Wins

A passenger types into the route search, and the page sends a request on every keystroke. Answers come back in any order. Today's page aborts the old request when it sends a new one, so a slow answer can never replace a newer one. Work out what that page shows, and what a page *without* aborting would end up showing.

Request `i` (numbered from 1) is sent at time `s_i` and its answer arrives at `s_i + d_i`.

- **With abort:** sending request `i+1` aborts request `i`. Request `i`'s answer is shown only if it arrives strictly before `s_(i+1)`. If it arrives at the same moment or later, it was aborted. The last request is never aborted.
- **Without abort:** every answer is shown when it arrives, replacing whatever was there. The page ends showing the answer that arrives last. If two arrive at the same moment, the higher-numbered one is drawn last.

**Input**

- Line 1: `n` (1 ≤ n ≤ 15 000)
- Next `n` lines: `s_i d_i` (0 ≤ s_i ≤ 10^9, with `s_i` strictly increasing; 0 ≤ d_i ≤ 10^6)

**Output**

- Line 1: `shown` followed by the requests shown with abort, in order
- Line 2: `without abort` followed by the request left on screen at the end without abort

**Example**

```
Input        Output
4            shown 2 4
0 300        without abort 3
100 50
200 500
250 100
```

Request 1 would arrive at 300, after request 2 was sent at 100: aborted. Request 2 arrives at 150, before 200: shown. Request 3 would arrive at 700, after 250: aborted. Request 4 is the last: shown. Without abort, request 3's answer lands last, at 700, and the passenger sees results for the wrong text.

**Limits:** 2 seconds per test case.

**Go tip:** read all the pairs into a slice first, because whether request `i` is shown depends on `s_(i+1)`. Collect the shown numbers as strings and print them with `strings.Join`.

Judge it from the repo root with `go run ./tools/kata 68`.
