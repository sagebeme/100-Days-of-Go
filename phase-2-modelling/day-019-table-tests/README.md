# Day 19: Table-Driven Tests

## The brief

A chama treasurer types contributions into a form: `500`, `Ksh 1,250`, `kes1250.50`, and sometimes `12,50` or `five hundred`. The app needs a parser that reads every sensible way of writing an amount and rejects the rest with a clear reason. A parser like this has dozens of edge cases, and it is exactly the kind of code that breaks quietly when someone "just tidies it up".

Today's code is small: `ParseAmount` and `FormatAmount`. The real subject is the test file next to it. It's the biggest you've seen so far, and you'll read it, run parts of it, measure it and add to it.

```
ParseAmount("Ksh1,250.50")   →  125050, nil      (cents)
ParseAmount("  kes 75 ")     →  7500, nil
ParseAmount("1250.5")        →  125050, nil
ParseAmount("12,50")         →  0, not an amount: "12,50"
ParseAmount("Ksh")           →  0, empty amount: "Ksh"
ParseAmount("1000000000")    →  0, amount too large: "1000000000"

FormatAmount(125050)         →  "Ksh1,250.50"
FormatAmount(-20000)         →  "-Ksh200.00"
```

Money is kept in **cents** as an `int`. Floats can't hold `0.1` exactly, so `0.1 + 0.2` isn't `0.3` in a `float64`. Whole numbers of cents never round.

## What you'll use

You've had table tests since [Day 1](../../phase-1-terminal/day-001-mpesa-sms/). Today you meet the rest of the `testing` toolkit:

- **Subtests**: `t.Run(name, func(t *testing.T) { ... })` runs each row as its own test, with its own name and its own pass or fail. A `t.Fatalf` inside a subtest stops only that row.
- **Running one subtest**: `-run` takes a pattern, with `/` between the test and the subtest. Spaces in names become `_`.

```bash
go test -v -run 'TestParseAmountErrors/comma_group' ./phase-2-modelling/day-019-table-tests/starter
```

- **Test helpers**: a function that checks something and calls `t.Errorf`. Put `t.Helper()` as its first line, so a failure reports the line that *called* the helper.
- **`errors.Is` in tests**: the error tests check *which* error came back, not its text. The message can be improved later without breaking a single test.
- **Coverage**: which lines your tests ran.

```bash
go test -cover ./phase-2-modelling/day-019-table-tests/starter
go test -coverprofile=cover.out ./phase-2-modelling/day-019-table-tests/starter
go tool cover -func=cover.out   # coverage per function
go tool cover -html=cover.out   # opens your code in the browser, untested lines in red
```

`cover.out` is in `.gitignore`. Don't commit it.

- **Fuzz tests** (a taste): `FuzzRoundTrip` takes seed inputs with `f.Add`. A plain `go test` runs only the seeds. `go test -fuzz=FuzzRoundTrip -fuzztime=10s` makes up thousands of new inputs and saves any that fail in `testdata/fuzz/`, so they become permanent test cases.

## Why it matters at work

In code review, "where's the test for that?" is the most common comment there is. A table makes the answer cheap: a new case is one line. Teams ask for a failing row with every bug fix, so the bug can't come back unnoticed. Subtest names show up in CI logs, so `TestParseAmountErrors/comma_group_too_short` tells a teammate what broke without opening the file. Coverage is a guide, not a goal. 100% means every line ran, not that every case was checked. Use the red lines in the HTML report to find branches nobody thought about, and don't write tests just to colour them green.

## Steps

1. Read `starter/amount_test.go` from top to bottom before you write any code. Find the helper, the two tables, `t.Run`, the round-trip test and the fuzz test.
2. Run the tests with `-v`. Every row fails, and each failure names its row.
3. Write `FormatAmount` first, since it's the smaller one: split the sign off, then `cents/100` and `cents%100`. [Day 13](../../phase-1-terminal/day-013-package-it-up/)'s comma code works here. `%02d` pads the cents to two digits. Run just its tests: `-run TestFormatAmount`.
4. Write `ParseAmount` one subtest at a time. Pick the first failing row, make it pass, run again. A good order: trim and strip the prefix, then split on `.` with `strings.Cut`, then check the whole part (plain digits, or comma groups with `strings.Split`), then the cents.
5. Wrap every error with the input: `fmt.Errorf("%w: %q", ErrFormat, s)`. The tests use `errors.Is`, so any message with `%w` passes.
6. When everything passes, run with `-cover`. Then open the HTML report and look for any red lines.
7. Add a row to each table that you think might break your code, for example `"Ksh 0.5"` or `"1,000,000"`. If it fails, decide whether the test or the code is wrong.
8. Run the fuzz test for ten seconds: `go test -run XXX -fuzz=FuzzRoundTrip -fuzztime=10s ./phase-2-modelling/day-019-table-tests/starter`. (`-run XXX` matches no normal tests, so only the fuzzer runs.)
9. Run the whole suite from the repo root:

```bash
go test ./phase-2-modelling/day-019-table-tests/starter
```

## The kata: Coverage Report

Your own mini `go test -cover`: given the blocks of lines each test ran, count the covered lines, print the percentage and list the gaps. It needs an idea you'll meet properly in Phase 3: a difference array, or sorting and merging ranges. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 19
```

## When you're stuck

- **A failure points inside `assertCents`, not at the row**: the helper is missing `t.Helper()`.
- **`-run 'TestParseAmount/commas and cents'` runs nothing**: subtest names use `_` for spaces: `commas_and_cents`.
- **`TestParseAmount/one_decimal_digit` gives 125005**: `"5"` after the dot means 50 cents, not 5. Pad one digit to two before you convert.
- **`"12,50"` parses as 1250**: you removed the commas before checking them. Check that every group after the first has exactly 3 digits, then remove them.
- **`"Ksh"` gives ErrFormat instead of ErrEmpty**: check for an empty string *after* removing the prefix and the spaces.
- **The fuzzer reports a failure**: it saved the input in `testdata/fuzz/FuzzRoundTrip/`. Run `go test` again and it replays that input, so you can debug it. Delete that folder once it's fixed, if you're not keeping the case.
- **`FormatAmount(5)` gives `Ksh0.5`**: use `%02d` for the cents.
- **Still stuck?** Open `solution/amount.go`, read it once, close it, then write your own version from memory.
