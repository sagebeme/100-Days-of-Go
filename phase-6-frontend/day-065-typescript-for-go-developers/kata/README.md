# Kata: Payment Events

A matatu SACCO's till logs every fare event on its own line. The first word says what kind of event it is, and the kind decides what comes after it. It's today's TypeScript union, as text:

| Line | Meaning |
|---|---|
| `mpesa AMOUNT PHONE` | a fare paid by M-Pesa |
| `cash AMOUNT` | a fare paid in cash |
| `card AMOUNT LAST4` | a fare paid by card |
| `refund AMOUNT` | money given back |
| `failed REASON` | a payment that didn't go through (one word, no amount) |

Add it all up.

**Input**

- Line 1: `n`, the number of events (0 ≤ n ≤ 10 000)
- Next `n` lines: one event each, in the forms above (1 ≤ AMOUNT ≤ 1 000 000)

**Output**

Six lines: the M-Pesa total, the cash total, the card total, the refund total, `net` (the three payment totals minus refunds, which can be negative) and the number of failed payments.

**Example**

```
Input                       Output
5                           mpesa 100
mpesa 100 0712345678        cash 150
cash 150                    card 100
card 100 4242               refunds 50
refund 50                   net 300
failed timeout              failed 1
```

**Limits:** 2 seconds per test case.

**Go tip:** `fmt.Fscan` reads one word at a time, so read the kind first and then `switch` on it to decide how many more words to read. That `switch` is the Go version of narrowing on `kind` in TypeScript:

```go
var kind string
fmt.Fscan(in, &kind)
switch kind {
case "mpesa":
	var amount int
	var phone string
	fmt.Fscan(in, &amount, &phone)
}
```

Judge it from the repo root with `go run ./tools/kata 65`.
