# Kata: Phone Numbers

The booking form takes phone numbers however people type them: `0712 345 678`, `+254712345678`, `0110-123-456`. The SMS gateway wants one form only: `254` and nine digits. Clean up a list of numbers, the same way as today's `normalisePhone`.

Remove every space and dash. Then the number is valid if it is either:

- `0`, then `7` or `1`, then 8 more digits (10 digits in all), or
- `254` or `+254`, then `7` or `1`, then 8 more digits.

A valid number becomes `254` followed by its last nine digits. Anything else is `INVALID`.

**Input**

- Line 1: `n` (1 ≤ n ≤ 10 000)
- Next `n` lines: one phone number each, as typed. A line may contain spaces.

**Output**

- `n` lines: each number cleaned up, or `INVALID`
- Then `unique K`: how many different valid numbers there were

**Example**

```
Input            Output
4                254712345678
0712 345 678     254712345678
+254712345678    INVALID
0812345678       254110123456
0110-123-456     unique 2
```

**Limits:** 2 seconds per test case.

**Go tip:** the numbers contain spaces, so `fmt.Fscan` would split them. Read whole lines with `bufio.Scanner` and `scanner.Text()`. `strings.NewReplacer(" ", "", "-", "")` strips both in one pass, and a `map[string]bool` counts the unique ones.

Judge it from the repo root with `go run ./tools/kata 69`.
