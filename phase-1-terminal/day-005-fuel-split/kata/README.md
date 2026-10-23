# Kata: Fuliza Check

You have a week of M-Pesa balances. Thanks to Fuliza, some of them are below zero. Print the lowest and the highest balance.

**Input**

- Line 1: `n`, the number of balances (1 ≤ n ≤ 200 000)
- Line 2: `n` whole numbers, the balances in shillings, separated by spaces (-1 000 000 000 ≤ balance ≤ 1 000 000 000)

**Output**

One line: the lowest balance, a space, then the highest balance.

**Example**

```
Input                     Output
5                         -250 1200
300 -250 1200 0 45
```

**Limits:** 2 seconds per test case.

**Go tip**

Start both the lowest and the highest at the first number, not at 0. If every balance is negative, a highest that starts at 0 would be wrong. A small function that returns two values keeps the loop tidy:

```go
func update(lo, hi, x int) (int, int) {
	if x < lo {
		lo = x
	}
	if x > hi {
		hi = x
	}
	return lo, hi
}

lo, hi = update(lo, hi, x)
```

Judge it from the repo root with `go run ./tools/kata 5`.
