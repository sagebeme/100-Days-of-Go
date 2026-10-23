# Kata: M-Pesa Code Checker

A transaction code is valid when all of these are true:

- it is exactly 10 characters long
- every character is a capital letter `A`–`Z` or a digit `0`–`9`
- the first character is a letter
- it contains at least one digit

Check each line and print whether it is valid.

**Input**

- Line 1: `n`, the number of codes (1 ≤ n ≤ 10 000)
- The next `n` lines: one code each. A line has at most 30 printable ASCII characters and can be empty or contain spaces.

**Output**

For each line, `VALID` or `INVALID`.

**Example**

```
Input           Output
4               VALID
QFT1ABC2DE      INVALID
qft1abc2de      INVALID
QFTXABCYDE      INVALID
1FT1ABC2DE
```

**Limits:** 2 seconds per test case.

**Go tip**

Read whole lines with a `bufio.Scanner`, since a line might contain a space. Then check the bytes with a loop:

```go
for i := 0; i < len(code); i++ {
	c := code[i]
	isLetter := c >= 'A' && c <= 'Z'
	isDigit := c >= '0' && c <= '9'
	_, _ = isLetter, isDigit
}
```

Print with a `bufio.Writer` and `defer out.Flush()`. Printing 10 000 lines straight to `os.Stdout` works, but buffering is the habit that keeps you fast on bigger tests.

Judge it from the repo root with `go run ./tools/kata 9`.
