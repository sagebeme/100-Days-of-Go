# Kata: Transfer Log

A small money-transfer service processes a day's transfers in order and logs each one as `OK` or as `FAILED` with a reason code. Write the processor.

For each transfer `code from to amount`, check these in order and stop at the first that applies:

1. `amount` is below 1: `INVALID_AMOUNT`
2. `from` or `to` isn't an account: `UNKNOWN_ACCOUNT`
3. `from` and `to` are the same account: `SAME_ACCOUNT`
4. `from` has less than `amount`: `INSUFFICIENT_FUNDS`

Otherwise move the money and log `OK`. A failed transfer changes nothing.

**Input**

- Line 1: `n`, the number of accounts (1 ≤ n ≤ 1000)
- The next `n` lines: an account name and its balance (0 ≤ balance ≤ 10^9). Names are unique, 1 to 20 lowercase letters and digits.
- Then `q`, the number of transfers (1 ≤ q ≤ 100 000)
- The next `q` lines: a code, the sender, the receiver and the amount (−10^9 ≤ amount ≤ 10^9)

**Output**

- One line per transfer: the code and `OK`, or the code, `FAILED` and the reason
- Then one line per account, in input order: its name and final balance

**Example**

```
Input                      Output
3                          QFT1 OK
amina 5000                 QFT2 FAILED INSUFFICIENT_FUNDS
kip 200                    QFT3 FAILED UNKNOWN_ACCOUNT
duka 0                     QFT4 FAILED INVALID_AMOUNT
5                          QFT5 OK
QFT1 amina duka 1200       amina 3800
QFT2 kip duka 500          kip 1400
QFT3 amina wanjiku 100     duka 0
QFT4 duka kip 0
QFT5 duka kip 1200
```

**Limits:** 2 seconds per test case. Totals stay below 2^62, so `int` is enough.

**Go tip**

This is today's project in miniature. A function that returns an `error` keeps the order of the checks in one place, and sentinel errors carry the reason codes:

```go
var errSame = errors.New("SAME_ACCOUNT")

func check(...) error {
	if from == to {
		return errSame
	}
	...
	return nil
}
```

`err.Error()` then gives you the text to print. Use a `bufio.Writer` for the output, and `defer out.Flush()`.

Judge it from the repo root with `go run ./tools/kata 17`.
