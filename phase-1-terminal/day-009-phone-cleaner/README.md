# Day 9: Phone Number Cleaner

## The brief

A chama in Nakuru keeps its members' phone numbers in a shared spreadsheet. Everyone typed their own: some with spaces, some with +254, some with dashes, a few with typos. The treasurer wants to send M-Pesa payouts, and the payment API only takes numbers like `254712345678`. Write a cleaner that fixes the good numbers and refuses the bad ones with a clear error.

```
Clean("0712 345 678")    →  "254712345678", nil
Clean("+254712345678")   →  "254712345678", nil
Clean("0110-123-456")    →  "254110123456", nil
Clean("0712a45678")      →  "", invalid character 'a' in "0712a45678"
Clean("0201234567")      →  "", "0201234567" isn't a mobile number: it should start 07 or 01
Mask("+254712345678")    →  "0712***678", nil
SameNumber("0712 345 678", "+254712345678")  →  true
```

A wrong number here sends someone's money to a stranger. So the rule is strict: when in doubt, return an error.

## What you'll use

- Indexing a string gives a `byte`: `s[0]` is the first byte. Phone numbers are plain ASCII, so one byte is one character.
- Comparing bytes with character literals: `c >= '0' && c <= '9'` is true for a digit.
- A `switch` with no value, where each `case` is a condition.
- `strings.TrimSpace` to drop spaces at both ends, and `strings.HasPrefix(s, "254")`.
- `errors.New` and `fmt.Errorf` from Day 5. `%q` prints a value in quotes, which makes errors easy to read.

```go
s := "07a1"
for i := 0; i < len(s); i++ {
	c := s[i]
	if c < '0' || c > '9' {
		fmt.Printf("%q at index %d isn't a digit\n", c, i) // 'a' at index 2 isn't a digit
	}
}
```

## Steps

1. Open `starter/phone.go`. Import `"errors"`, `"fmt"` and `"strings"` as you need them.
2. In `Clean`, trim the spaces from both ends. If nothing is left, return an error.
3. Walk the string byte by byte. Collect digits into a string called `digits`. Skip spaces and hyphens. Allow `+` only at index 0, and remember that you saw it. Return an error for any other byte, and say which character it was.
4. Work out the 9-digit local part. 12 digits starting `254`: drop the `254`. Otherwise, if there was a `+`, that's an error. 10 digits starting `0`: drop the `0`. Exactly 9 digits: keep them all. Any other count is an error.
5. The local part must start with `7` or `1`. Anything else, like a `020` landline, is an error. Return `"254"` plus the local part.
6. `Mask`: call `Clean`. If it fails, return its error. Otherwise build `"0"`, the first three local digits, `"***"`, then the last three.
7. `SameNumber`: clean both. If either fails, return `false`. Otherwise compare them.
8. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-009-phone-cleaner/starter
```

## The kata: M-Pesa Code Checker

Every M-Pesa transaction has a 10-character code like `QFT1ABC2DE`. Read a list of codes and say which ones have the right shape, checking each one character by character. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 9
```

## When you're stuck

- **`invalid operation: c >= "0" (mismatched types byte and untyped string)`**: `"0"` is a string. Use single quotes, `'0'`, for a single character.
- **`cannot use c (variable of type byte) as string value`**: to add a byte to a string, convert it: `digits += string(c)`.
- **Your error message prints `97` instead of `'a'`**: use `%q`, not `%d`, for the byte.
- **`index out of range [0] with length 0`**: you read `local[0]` before checking that there were any digits. Check the length first.
- **`+0712345678` is accepted**: after you see a `+`, only the 12-digit `254` form is allowed. Check that case before the `0` case.
- **Still stuck?** Open `solution/phone.go`, read it once, close it, then write your own version from memory.
