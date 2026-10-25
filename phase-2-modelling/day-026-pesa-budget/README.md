# Day 26: Pesa Budget CLI

## The brief

Kevin has his M-Pesa statements as CSV files now (Day 24), and a budget in his head: Ksh3,000 a month for food, Ksh500 for the matatu. What he doesn't have is a way to see, in one command, whether April went to plan. Build `pesabudget`.

It reads one or more statement files, sorts every completed payment into a category by matching words in its details, compares each category with its budget and prints a report. It's configured the way real tools are: env variables for the settings you keep, flags for the ones you change per run, with flags winning. With `-strict` it exits with a special code when you're over budget, so a script or a cron job can nag you.

This is the Phase 2 capstone. There is no walkthrough today: you get the brief, the requirements, the tests and some hints. Everything you need is from Days 15 to 25.

```
$ export PESA_BUDGETS=food=3000,transport=500,fun=1000
$ pesabudget -month 2026-04 -strict testdata/march.csv testdata/april.csv
Pesa Budget: 2026-04
CATEGORY              SPENT         BUDGET           LEFT
food            Ksh4,100.00    Ksh3,000.00   -Ksh1,100.00 OVER
bills           Ksh1,600.00              -              -
airtime           Ksh250.00              -              -
transport         Ksh240.00      Ksh500.00      Ksh260.00
fees               Ksh10.00              -              -
fun                 Ksh0.00    Ksh1,000.00    Ksh1,000.00
TOTAL           Ksh6,200.00    Ksh4,500.00      Ksh160.00
money in: Ksh8,000.00
$ echo $?
3

$ pesabudget -month 2026-4 testdata/april.csv
pesabudget: month "2026-4" must look like 2026-04          (exit code 2)

$ PESA_RULES="NYAMA CHOMA=eating-out" pesabudget -month 2026-04 testdata/april.csv
...
eating-out        Ksh950.00              -              -
...
```

The two statements in `starter/testdata/` are made up, in the Day 24 format. `april.csv` repeats the last row of `march.csv`, as real exports that overlap do.

## What you'll use

- **Day 24's statement parser**, as `ParseStatement`. Copy your own code in; it's yours.
- **Structs, methods and slices** (Day 15) for `Rule`, `Line` and `Summary`, and a method `Summary.Over`.
- **Errors** (Day 17) that name the setting or the file and line that's wrong.
- **Table tests** (Day 19). The test file is long; read it before you start, because it is most of the spec.
- **Config from the environment** (Day 21): a `LookupFunc` instead of `os.LookupEnv`, so the tests never touch your real environment. [reference/env-and-config.md](../../reference/env-and-config.md) has the patterns.
- **Flags and exit codes** (Day 22): one `FlagSet`, `fs.Visit` to know which flags were really given, `Run(args, lookup, stdout, stderr) int`.
- **Files** (Day 23), opened and closed with `defer`.
- **`fmt` widths** (Day 11): `%-12s` pads on the right, `%14s` on the left.

## Why it matters at work

This is the shape of most small internal tools: read some files, apply rules, print a report, return an exit code a machine can act on. The engineering habits are what turn a script into a tool your team trusts. Settings are checked before any work starts, so a typo in `PESA_BUDGETS` fails in the first millisecond with a message that says where it came from. There's one clear order of precedence, flag over env over default. There's no hidden global state: `main` is one line, and everything else can be tested. Results go to stdout and problems to stderr, so `pesabudget ... > april.txt` gives you a clean file. And the money is never a `float64`.

## Steps

Requirements. Meet them in any order; the tests check each one.

1. `ParseStatement(r)` reads a statement exactly as Day 24's `Parse` does: columns by name, byte order mark ignored, amounts in cents, `Withdrawn` stored as a positive number, errors starting `line N:`.
2. `Categorise(details, rules)` returns the category of the first rule whose keyword appears in the details, ignoring case, or `"other"`. `DefaultRules` is given, and its order matters.
3. `ParseRules` reads `KEYWORD=category;KEYWORD=category` and `ParseBudgets` reads `category=shillings,category=shillings`. Both trim spaces and lower-case the category, return empty (non-nil) results for `""`, and return the errors listed in their doc comments. Budgets are stored in cents.
4. `Summarise(txs, rules, budgets, month)`:
   - counts only `Completed` transactions in `month` (`"2026-04"`), or in every month when `month` is `""`
   - counts a receipt number only the first time it appears
   - adds withdrawals per category, and paid-in money into `PaidIn`
   - gives every budgeted category a `Line`, even with nothing spent
   - sorts lines by spent, highest first, then by category name
5. `Summary.Over()` is true when any budgeted category spent more than its budget.
6. `Report(s)` prints the table exactly as in the brief and the doc comment: the title, the header row, one row per line with `-` where there's no budget and ` OVER` after an overspent row, the `TOTAL` row, and the `money in:` line. `Ksh(cents)` formats every amount.
7. `Run(args, lookup, stdout, stderr)`:
   - flags `-month`, `-budgets` and `-strict`, falling back to `PESA_MONTH`, `PESA_BUDGETS` and `PESA_STRICT`, then to the defaults (every month, no budgets, not strict). An env variable set to `""` counts as unset.
   - `PESA_RULES` adds rules that are tried before `DefaultRules`.
   - the statement files are the arguments after the flags. At least one is required. All of them are summarised together.
   - exit codes: `0` success, `1` a file can't be opened or parsed, `2` a bad flag, setting or missing file name, `3` with `-strict` when over budget (the report is still printed).
   - every problem is one line on stderr starting with `pesabudget: `, naming its source (`-budgets:`, `PESA_BUDGETS:`, or the file path). Nothing goes to stdout when it fails.
8. `starter/cmd/pesabudget/main.go` is written: one line that calls `Run` with `os.Args[1:]` and `os.LookupEnv`.
9. Run the tests from the repo root, then the real thing:

```bash
go test ./phase-2-modelling/day-026-pesa-budget/starter
PESA_BUDGETS=food=3000,transport=500 go run ./phase-2-modelling/day-026-pesa-budget/starter/cmd/pesabudget \
    -month 2026-04 phase-2-modelling/day-026-pesa-budget/starter/testdata/*.csv
```

When it works, make it yours. Add rules for the places you actually spend, export a real statement of your own (keep it out of git: `*.csv` files with your name and phone number don't belong in a repo), and see where your money goes.

## The kata: Budget Alerts

Instead of a report at the end of the month, warn the moment a category goes over its budget: once per category per month. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 26
```

## When you're stuck

- **Where do I start?** Read the test file top to bottom. Then do the pure functions first (`Ksh`, `Categorise`, `ParseBudgets`, `ParseRules`), then `ParseStatement`, `Summarise` and `Report`, and `Run` last. Run one test at a time with `go test -run TestKsh ./phase-2-modelling/day-026-pesa-budget/starter`.
- **`Customer Transfer of Funds Charge` comes out as `sent`**: you sorted or reversed the rules. The first matching rule wins, and `Charge` is first on purpose.
- **Food for all months is Ksh6,965.50, not Ksh6,325.50**: the March row repeated in `april.csv` counted twice. Keep a `map[string]bool` of receipts you've seen.
- **April includes March spending**: compare `tx.Time.Format("2006-01")` with the month.
- **`fun` is missing from the report**: add a zero line for every budgeted category nobody spent on.
- **The rows come out in a different order on each run**: you ranged over a map. Build the slice, then `sort.Slice` it with both the spent and the name rules.
- **The columns are one space off**: the format is `%-12s %14s %14s %14s`, with single spaces between, and ` OVER` added right after the last column.
- **`-month` given, but `PESA_MONTH` still wins**: read env values only for flags that weren't set. `fs.Visit` visits only the flags that were given.
- **`-strict=false` doesn't beat `PESA_STRICT=true`**: same cause. `fs.Visit` counts `-strict=false` as set.
- **A failed run still printed a report**: check every setting and read every file before you print anything.
- **Still stuck?** Open the files in `solution/`, read them once, close them, then write your own version from memory.
