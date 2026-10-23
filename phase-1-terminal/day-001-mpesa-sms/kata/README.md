# Kata: Till Total

A duka's till adds up everything a customer buys. Read the prices and print the total.

**Input**

- Line 1: `n`, the number of items (0 ≤ n ≤ 200 000)
- Line 2: `n` prices in shillings, separated by spaces (1 ≤ price ≤ 1 000 000 000)

**Output**

One line: the total.

**Example**

```
Input           Output
3               200
50 120 30
```

**Limits:** 2 seconds per test case. The biggest totals don't fit in 32 bits; Go's `int` is 64 bits on every computer you'll use, so it's fine.

**Reading input in Go**

```go
in := bufio.NewReader(os.Stdin)
var n int
fmt.Fscan(in, &n) // reads the next number, skipping spaces and newlines
```

Judge it from the repo root with `go run ./tools/kata 1`. Each test case shows `PASS` or `FAIL` and how long it took.
