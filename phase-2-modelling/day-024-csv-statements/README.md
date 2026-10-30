# Day 24: CSV Statements

## The brief

Kevin wants to know where his money goes every month. Safaricom will send him a full M-Pesa statement, and once it's turned into a spreadsheet he has rows like these:

```
Receipt No.,Completion Time,Details,Transaction Status,Paid In,Withdrawn,Balance
SC71ABCD12,2026-03-01 08:15:42,Funds received from 0722***111 - JOSEPH OTIENO,Completed,"5,000.00",,"5,250.00"
SC71ABCD13,2026-03-01 12:02:10,Merchant Payment to 512345 - SOKO BORA SUPERMARKET,Completed,,-1235.50,"4,014.50"
SC74DEFG41,2026-03-04 09:10:00,Merchant Payment to 512345 - SOKO BORA SUPERMARKET,Failed,,-2000.00,"2,696.50"
SC74DEFG42,2026-03-04 17:20:31,"Merchant Payment to 600100 - MAMA PIMA GROCERS, KASARANI",Completed,,-350.00,"2,346.50"
```

Write the parser that turns that file into Go values, with errors that say which line is broken. Then add up what came in and went out, and who he paid.

```
txs, err := Parse(file)       →  12 transactions, nil
txs[1].Withdrawn              →  123550         (cents)
txs[1].Time                   →  2026-03-01 12:02:10 +0300 EAT
Totals(txs)                   →  575000, 376350 (the failed Ksh2,000 doesn't count)
Counterparty(txs[1].Details)  →  "SOKO BORA SUPERMARKET"
SpentWith(txs)["SOKO BORA SUPERMARKET"]  →  187550
```

*Safaricom sends full statements as PDFs. The CSV layout here is modelled on a statement's columns, and every name, number and receipt in `testdata/statement.csv` is made up.*

## What you'll use

New today: **`encoding/csv`**.

- `csv.NewReader(r)` reads from any `io.Reader`: a file, or `strings.NewReader` in a test. `Read()` returns one record as a `[]string`, and `io.EOF` when the file is done.
- It handles the hard parts of CSV for you: `"MAMA PIMA GROCERS, KASARANI"` is one cell, not two; `""` inside quotes is one `"`; a quoted cell may even span lines. Never split CSV lines with `strings.Split(line, ",")`.
- By default every record must have as many cells as the first one. If not, `Read` returns a `*csv.ParseError` whose message says `record on line 3: wrong number of fields`. Blank lines are skipped.
- `r.FieldPos(0)` returns the line and column where the current record's first cell started, so your own errors can say `line 7:` too.
- **Find columns by name, not position.** Read the header row into a `map[string]int` from name to index. Then a file with the columns in another order, or an extra column, still works.
- **The byte order mark.** Excel and some exporters put the invisible character `U+FEFF` at the very start of a file. Your first header cell is then `"\ufeffReceipt No."`, and a lookup for `"Receipt No."` fails. Strip it with `strings.TrimPrefix(name, "\ufeff")`.
- **Times with a zone.** `time.ParseInLocation(layout, s, EAT)` reads a time that has no zone written in it, in the zone you give. `time.FixedZone("EAT", 3*60*60)` is UTC+3. Kenya doesn't use daylight saving, so a fixed zone is correct. Comparing times uses `t.Equal(u)`, not `==`.
- Money again in **cents**, as on [Day 19](../day-019-table-tests/): `"1,235.50"` becomes `123550`.

## Steps

1. Open `testdata/statement.csv` in the starter folder and read it. Notice the quoted amounts, the minus signs in `Withdrawn`, the failed payment, and the comma inside one `Details` cell. Look at the first bytes with `head -c 20 starter/testdata/statement.csv | od -c`, and you'll see the byte order mark: `357 273 277`.
2. Open `starter/statement.go`. Write `Parse` in stages. First read the header and build the `map[string]int`, stripping the byte order mark from the first cell and the spaces from every cell. Check that every name in `Columns` is there.
3. Loop over the records with `Read` until `io.EOF`. For each one, get the line with `cr.FieldPos(0)` and parse the cells. A small helper `parseRow` and an amount helper `cents(s string) (int, error)` keep `Parse` readable.
4. In the amount helper, remove the commas with `strings.ReplaceAll`, take off a leading `-`, split on `.` with `strings.Cut`, and pad the cents to two digits. An empty cell is 0.
5. Store `Withdrawn` as a positive number. A positive value in the file's `Withdrawn` column, or a negative `Paid In`, is an error.
6. Write `Totals` (completed rows only), `Counterparty` (`strings.Cut(details, " - ")`) and `SpentWith`.
7. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-024-csv-statements/starter
```

## The kata: Column Sum

Sum one column of a CSV with quoted cells, embedded commas, doubled quotes and newlines inside cells: exactly the cases that break `strings.Split`. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 24
```

## When you're stuck

- **`line 1: missing column "Receipt No."`, though it's right there**: the byte order mark is stuck to the first header cell. Print it with `%q` to see `"\ufeffReceipt No."`.
- **`record on line 9: wrong number of fields`**: a row has a comma that isn't inside quotes. The reader is right; the file is broken. Your error should pass this on, not hide it.
- **`MAMA PIMA GROCERS, KASARANI` comes out as two cells**: you split the line yourself. Let `csv.Reader` do it.
- **The first transaction's time is three hours off**: you used `time.Parse`, which assumes UTC. Use `time.ParseInLocation` with `EAT`.
- **`TestTotals` wants 376350 and you get 576350**: you counted the failed payment. Check `Status == "Completed"`.
- **`123549` instead of `123550`**: you went through `float64` (`1235.50 * 100`). Parse the digits as whole numbers.
- **`TestParseEmptyStatement` wants `[]Transaction{}`**: start with `txs := []Transaction{}`, so a statement with no rows isn't `nil`.
- **Still stuck?** Open `solution/statement.go`, read it once, close it, then write your own version from memory.
