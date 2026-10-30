# Day 1: M-Pesa SMS Builder

## The brief

Send money on M-Pesa and a text arrives a second later: a transaction code, the word "Confirmed", the amount, and who got it. Write a function that builds that confirmation from its three parts.

```
BuildConfirmation("QFT1ABC2DE", 500, "Amina Wanjiru")
  →  "QFT1ABC2DE Confirmed. Ksh500.00 sent to Amina Wanjiru."

BuildConfirmation("RKD4XY7Z1P", 1250, "Kip")
  →  "RKD4XY7Z1P Confirmed. Ksh1250.00 sent to Kip."
```

*This is a simplified version of the real text. Real messages also put commas in big amounts and add the phone number, the date and time, and your new balance. You'll format numbers properly on [Day 11](../day-011-receipt-printer/).*

## What you'll use

- `const` for a value that never changes, like the word `"Confirmed."`
- `:=` to declare a variable and give it a value in one go, and `=` to change it later
- `+` to join strings, and `fmt.Sprintf` to build a string from a template: `%s` stands for a string and `%d` for a whole number

```go
greeting := "Habari"
greeting = greeting + ", Wanjiku"           // "Habari, Wanjiku"
line := fmt.Sprintf("%s owes %d", "Kip", 40) // "Kip owes 40"
```

Go is strict in two ways you'll meet today. An `int` can't be glued onto a string with `+` (that's what `Sprintf` is for), and a variable you declare but never use is a compile error, not a warning.

## Steps

1. Open `starter/sms.go`. The function's name and parameters are already there; don't change them, because the tests call it exactly like that.
2. Declare two constants: `status` holding `"Confirmed."` and `currency` holding `"Ksh"`.
3. Declare `message := code + " " + status`.
4. Add the rest to the end of `message`: a space, the currency stuck to the amount, `.00`, then ` sent to `, the name and a full stop. `fmt.Sprintf(" %s%d.00 sent to %s.", ...)` does it in one line. Add `import "fmt"` under the `package` line.
5. Return `message`.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-001-mpesa-sms/starter
```

`ok` means you're done. `FAIL` shows what your function returned next to what the test wanted.

## The kata: Till Total

Every day ends with a short puzzle in the format HackerRank uses: your program reads its input from the keyboard (standard input) and prints its answer. Instructions are in `kata/README.md`. Write it in `kata/main.go` and judge it with:

```bash
go run ./tools/kata 1
```

## When you're stuck

- **`undefined: fmt`**: add `import "fmt"` on the line after `package mpesasms`.
- **`declared and not used`**: Go refuses to compile a variable you never read. Use it, or delete it.
- **`invalid operation: mismatched types string and int`**: you tried `"Ksh" + amount`. Use `fmt.Sprintf` with `%d` for the number.
- **The test shows `Ksh 500.00` or `Ksh500`**: compare your spaces with the brief character by character. There's no space between `Ksh` and the amount, and `.00` is plain text after `%d`.
- **Still stuck?** Open `solution/sms.go`, read it once, close it, then write your own version from memory.
