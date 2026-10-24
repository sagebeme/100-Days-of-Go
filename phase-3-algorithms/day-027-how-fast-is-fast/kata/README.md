# Kata: Deposit Pairs

An M-Pesa agent in Kisumu writes the day's work as one long string: `D` for a deposit and `W` for a withdrawal, in the order they happened. The auditor wants to count every pair where a deposit came before a withdrawal (not necessarily right before it). Print that count.

**Input**

- Line 1: `n`, the length of the log (1 ≤ n ≤ 2 × 10^5)
- Line 2: the log, `n` characters, each `D` or `W`

**Output**

One line: the number of pairs (i, j) with i < j, `log[i]` = `D` and `log[j]` = `W`.

**Example**

```
Input           Output
4               3
DWDW
```

Number the characters from 1. The pairs are (1, 2), (1, 4) and (3, 4).

**Limits:** 2 seconds per test case.

**Read the constraints first.** n can be 2 × 10^5, so checking every pair is about 2 × 10^10 steps and will time out. One pass, O(n), is about 2 × 10^5 steps. The answer can be as big as 10^10, which doesn't fit in 32 bits. Go's `int` is 64 bits, so it's fine.

**Go tip**

The log is one 200 000-character line. `fmt.Fscan` on a `bufio.Reader` reads a whole word into a string, however long:

```go
in := bufio.NewReader(os.Stdin)
var n int
var log string
fmt.Fscan(in, &n, &log)
```

A `bufio.Scanner` with its default settings gives up on lines over 64 KB. If you use one, call `sc.Buffer` first (see [reference/big-o.md](../../../reference/big-o.md#fast-input-and-output)).

Judge it from the repo root with `go run ./tools/kata 27`.
