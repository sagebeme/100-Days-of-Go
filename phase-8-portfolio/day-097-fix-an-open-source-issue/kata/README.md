# Kata: Amount Parser

People type money in every way: `Ksh 1,200.5`, `kes99`, `-ksh 20.05`, `1234567`. Write the parser the `pesa` package should have had: read each line, and print the amount in one standard form, or `INVALID`. Parsing a small grammar exactly, edge cases and all, is a classic interview question (LeetCode's "valid number" is the hard version).

**The grammar**

After removing spaces at both ends of the line, an amount is, in order:

1. an optional `-`
2. an optional currency, `Ksh` or `KES` in any mix of upper and lower case, followed by any number of spaces (including none)
3. an optional `-`, but only if there was no `-` in step 1
4. the shillings: either plain digits (`1234567`), or 1 to 3 digits followed by groups of a comma and exactly 3 digits (`1,234,567`). At most 13 digits, not counting commas. Leading zeros are allowed.
5. optionally, a `.` followed by one or two digits. One digit is tenths: `.5` is 50 cents.

Anything else is `INVALID`: an empty line, `.5`, `5.`, `1,23`, `--1`, `USD 5`, `1 000`, three decimal places.

**Input**

- Line 1: `Q`, the number of lines that follow (1 ≤ Q ≤ 100 000)
- The next `Q` lines: one possible amount each, up to 40 characters. A line can be empty.

**Output**

For each line, the amount as `Ksh` and the shillings with commas every three digits, then `.` and exactly two digits of cents. A negative amount starts with `-Ksh`. Zero is never negative: `-0` prints `Ksh 0.00`. Otherwise print `INVALID`.

**Example**

```
Input             Output
7                 Ksh 1,200.50
Ksh 1,200.5       Ksh 99.00
KES 99            -Ksh 20.05
-ksh 20.05        INVALID
1,23              -Ksh 0.50
Ksh -0.5          INVALID
12.345            Ksh 1,234,567.00
  1234567
```

`1,23` has a group of two digits after the comma, and `12.345` has three decimal places.

**Limits:** 2 seconds per test case.

**Go tip**

Read lines with `bufio.Scanner`, not `fmt.Fscan`, because a line can be empty or hold spaces. Peel the grammar off one piece at a time with `strings.CutPrefix` and `strings.Cut`, checking each piece before moving on:

```go
whole, frac, hasPoint := strings.Cut(s, ".") // "1,200", "5", true
```

Keep money in an `int64` of cents. Never use `float64` for it: `0.1 + 0.2` is not `0.3`.

Judge it from the repo root with `go run ./tools/kata 97`.
