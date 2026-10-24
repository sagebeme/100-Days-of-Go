# Kata: Payment Router

A shop's till holds several ways to pay: M-Pesa wallets, bank cards and cash drawers. For each payment, the till picks the method with the **lowest fee** among the methods that can afford it. If two methods charge the same fee, it picks the one listed first. Then it takes the money.

The fees are the same as today's project:

| Method | Fee for paying `a` shillings | Can afford `a` when |
|---|---|---|
| `mpesa` | 0 up to 100, 7 up to 500, 13 up to 1 000, 33 up to 5 000, 57 above that | balance ≥ a + fee |
| `card` | 2% of `a`, rounded down | what's left of the limit ≥ a + fee |
| `cash` | 0 | cash ≥ a |

A payment takes `a + fee` from the method it uses.

**Input**

- Line 1: `n`, the number of methods (1 ≤ n ≤ 10)
- The next `n` lines: a type (`mpesa`, `card` or `cash`) and the money it holds in shillings: the balance, the card limit or the cash (0 ≤ value ≤ 10^9)
- Then `q`, the number of payments (1 ≤ q ≤ 100 000)
- The next `q` lines: one amount each (1 ≤ amount ≤ 1 000 000)

**Output**

- One line per payment: the number of the method that paid (counting from 1) and the fee, or `DECLINED` if no method can afford it
- A last line: what each method has left, in input order, separated by spaces

**Example**

```
Input          Output
3              1 0
cash 1000      2 13
mpesa 5000     3 6
card 5000      DECLINED
4              200 4187 4694
800
800
300
6000
```

The first Ksh800 goes on cash (free). The second can't: only Ksh200 is left. M-Pesa charges 13 and the card 16, so M-Pesa pays. For Ksh300 the card's fee is 6 and M-Pesa's is 7. Nobody can cover Ksh6000.

**Limits:** 2 seconds per test case.

**Go tip**

Give each kind of method its own type with a `fee` and a `canPay` method, and keep them all in one `[]method` slice of an interface. The loop over the payments then never needs to know which kind it's looking at.

```go
type method interface {
	fee(a int) int
	left() int
	take(n int)
}
```

Judge it from the repo root with `go run ./tools/kata 16`.
