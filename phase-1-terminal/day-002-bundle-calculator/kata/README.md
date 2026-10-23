# Kata: Data Diary

You write down how many megabytes you use each day. Print the total and the average per day.

**Input**

- Line 1: `n`, the number of days (1 ≤ n ≤ 200 000)
- Line 2: `n` whole numbers, the megabytes used each day, separated by spaces (0 ≤ mb ≤ 100 000)

**Output**

- Line 1: the total megabytes, as a whole number
- Line 2: the average per day, with exactly 2 decimal places

**Example**

```
Input           Output
3               700
200 0 500       233.33
```

**Limits:** 2 seconds per test case.

**Go tip**

The total is an `int`, but the average needs decimals. Convert both sides before you divide, then let `Printf` round it:

```go
average := float64(total) / float64(n)
fmt.Printf("%.2f\n", average) // 233.33
```

Judge it from the repo root with `go run ./tools/kata 2`.
