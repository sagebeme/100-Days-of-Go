# Kata: Statement Column

Print each amount with a comma between every group of three digits, right-aligned so that the last digits of every line sit in the same column. The column is exactly as wide as the widest formatted amount.

**Input**

- Line 1: `n`, the number of amounts (1 ≤ n ≤ 10 000)
- Line 2: `n` whole numbers separated by spaces (−10¹⁵ ≤ amount ≤ 10¹⁵). A negative amount keeps its minus sign in front: `-1,250`.

**Output**

`n` lines, one formatted amount per line, padded on the left with spaces.

**Example**

Input:

```
4
500 1250000 -30 7
```

Output. The widest amount is `1,250,000`, 9 characters, so every line is 9 characters wide:

```
      500
1,250,000
      -30
        7
```

**Limits:** 2 seconds per test case. 10¹⁵ fits easily in Go's 64-bit `int`.

**Go tip**

Format every amount first and keep the strings in a slice. Find the longest, then print them with a width taken from a variable, using `*`:

```go
fmt.Fprintf(out, "%*s\n", width, s) // right-aligns s in width characters
```

Judge it from the repo root with `go run ./tools/kata 11`.
