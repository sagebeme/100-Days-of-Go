# Kata: Budget Alerts

A budget app should warn you the moment you go over, not at the end of the month. You get monthly budgets per category and a list of spending in date order. Print an alert for each transaction that takes its category **over** its budget for that month, which means spending more than the budget. Being exactly on budget is fine.

Each category alerts at most once per calendar month, and its total starts again from 0 each month. A category with no budget never alerts.

**Input**

- Line 1: `b`, the number of budgets (0 ≤ b ≤ 100)
- The next `b` lines: a category and its monthly budget in shillings (0 ≤ budget ≤ 10^9)
- Then `n`, the number of transactions (1 ≤ n ≤ 200 000)
- The next `n` lines: a date `YYYY-MM-DD`, a category and an amount in shillings (0 ≤ amount ≤ 10^6), in date order

Categories are 1 to 20 lowercase letters.

**Output**

One line per alert, in input order: the date, the category, the month's total so far and the budget. If there are no alerts, print `no alerts`.

**Example**

```
Input                  Output
2                      2026-03-30 food 3100 3000
food 3000              2026-04-03 transport 540 500
transport 500
6
2026-03-28 food 2500
2026-03-30 food 600
2026-03-31 food 900
2026-04-01 food 3000
2026-04-02 transport 240
2026-04-03 transport 300
```

Food reaches 3100 on 30 March: alert. On 31 March it's still over, but it has already alerted this month. On 1 April it starts again and 3000 is exactly on budget.

**Limits:** 2 seconds per test case.

**Go tip**

The month is the first seven characters of the date, `date[:7]`. A struct makes a good map key when you need two things at once:

```go
type key struct{ month, category string }
spent := map[key]int{}
alerted := map[key]bool{}
```

Judge it from the repo root with `go run ./tools/kata 26`.
