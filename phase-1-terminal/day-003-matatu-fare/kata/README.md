# Kata: Stage Takings

A makanga collects the fare from every passenger. Children under 12 pay Ksh30, everyone from 12 to 64 pays Ksh60, and anyone 65 or older rides free. Read the ages and print the total collected.

**Input**

- Line 1: `n`, the number of passengers (0 ≤ n ≤ 200 000)
- Line 2: `n` ages, separated by spaces (0 ≤ age ≤ 120)

**Output**

One line: the total fare in shillings.

**Example**

```
Input             Output
5                 180
8 34 70 12 11
```

The 8-year-old and the 11-year-old pay 30 each. 34 and 12 pay 60 each. 70 rides free. 30 + 60 + 0 + 60 + 30 = 180.

**Limits:** 2 seconds per test case.

**Go tip**

Check the edges first. Is 12 a child or an adult? Is 65 a senior? Write the conditions so each age falls in exactly one band:

```go
switch {
case age < 12:
	total += 30
case age < 65:
	total += 60
}
```

Judge it from the repo root with `go run ./tools/kata 3`.
