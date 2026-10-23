# Kata: Dry Spell

A farmer writes down the rainfall in millimetres every day. A dry day is a day with 0 mm. Find the longest run of dry days in a row.

**Input**

- Line 1: `n`, the number of days (1 ≤ n ≤ 200 000)
- Line 2: `n` whole numbers, the rainfall each day in mm, separated by spaces (0 ≤ mm ≤ 500)

**Output**

One line: the length of the longest run of dry days. Print 0 if it rained every day.

**Example**

```
Input                   Output
8                       3
0 4 0 0 0 12 0 0
```

Days 3, 4 and 5 are dry in a row. Days 7 and 8 are only two.

**Limits:** 2 seconds per test case.

**Go tip**

Keep two accumulators: the streak you're in now and the best streak so far. Update the best every time the current one grows:

```go
if mm == 0 {
	current++
	if current > best {
		best = current
	}
} else {
	current = 0
}
```

Judge it from the repo root with `go run ./tools/kata 4`.
