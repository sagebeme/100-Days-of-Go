# Day 11: Receipt Printer

## The brief

Mama Njeri runs a duka in Kawangware. Her new thermal printer takes plain text, and she wants receipts that look like a supermarket's: amounts with commas, two decimal places, and columns that line up. Write the formatting functions and the receipt itself.

```
Commas(1250000)              →  "1,250,000"
Shillings(125000)            →  "Ksh1,250.00"
Shillings(5)                 →  "Ksh0.05"
Line("Unga 2kg", 2, 23000)   →  "Unga 2kg           2      Ksh460.00"

Receipt([]string{"Unga 2kg", "Sukari 1kg", "Milk 500ml"}, []int{2, 1, 3}, []int{23000, 18500, 6000})
```

prints:

```
Item        Qty  Total
Unga 2kg    2    Ksh460.00
Sukari 1kg  1    Ksh185.00
Milk 500ml  3    Ksh180.00
TOTAL            Ksh825.00
```

Prices are in cents: `23000` is Ksh230.00. Money is kept in whole numbers because a `float64` can't store `0.10` exactly, and a shop's totals must never be off by a cent.

## What you'll use

- Width and precision in `fmt` verbs. `%5d` pads a number to 5 characters on the left. `%-10s` pads a string to 10 characters on the right. `%.3s` keeps at most 3 characters. `%02d` pads with zeros.
- `strconv.Itoa(n)` turns an int into its digits as a string, so you can walk them.
- `/` and `%` on ints: `125050 / 100` is `1250` and `125050 % 100` is `50`.
- `text/tabwriter` lines up columns for you. Separate cells with `\t`, write the lines, then call `Flush`.

```go
fmt.Sprintf("[%5d]", 42)      // "[   42]"
fmt.Sprintf("[%-6s]", "Kip")  // "[Kip   ]"
fmt.Sprintf("[%.4s]", "Nairobi") // "[Nair]"
fmt.Sprintf("%02d", 5)        // "05"

var b strings.Builder
w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0) // pad columns with 2 spaces
fmt.Fprintln(w, "Kip\t40")
fmt.Fprintln(w, "Wanjiku\t1200")
w.Flush()
// b.String() is "Kip      40\nWanjiku  1200\n"
```

## Steps

1. Open `starter/receipt.go`. Import `"errors"`, `"fmt"`, `"strconv"`, `"strings"` and `"text/tabwriter"` as you use them.
2. `Commas`: remember the sign and make `n` positive. Turn it into digits with `strconv.Itoa`. Walk the digits and write a comma before digit `i` whenever `i > 0` and the number of digits left, `len(digits) - i`, divides by 3. Put the sign back in front.
3. `Shillings`: handle the sign the same way. The shillings are `cents / 100` and the cents are `cents % 100`. Build the result with `fmt.Sprintf("%sKsh%s.%02d", ...)`.
4. `Line`: one `fmt.Sprintf` with `%-16.16s`, `%3d` and `%14s`, separated by single spaces. The row total is `qty * priceCents`, formatted with `Shillings`.
5. `Receipt`: return an error if the three slices have different lengths. Make a `tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)` writing into a `strings.Builder`. Write the header `Item\tQty\tTotal`, then one line per item, then `TOTAL\t\t` and the grand total. Return an error for a quantity below 1 or a negative price. Call `w.Flush()` before you return `b.String()`.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-011-receipt-printer/starter
```

## The kata: Statement Column

An M-Pesa statement shows its amounts in one column, with commas, lined up on the right. Read a list of amounts and print them that way. Formatted output with exact widths is a common HackerRank task, and one stray space fails it. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 11
```

## When you're stuck

- **`Commas(1000)` gives `"1000"` or `",1,000"`**: print the comma before a digit only when `i > 0` and the digits left divide by 3.
- **`Shillings(5)` gives `"Ksh0.5"`**: use `%02d` for the cents, so 5 becomes `05`.
- **`-Ksh-1,250.50` or `Ksh-1,250.50`**: take the sign off first, format the positive amount, then put `-` in front of `Ksh`.
- **The receipt test shows no output at all**: tabwriter keeps everything until you call `w.Flush()`.
- **The columns are one space off**: the padding argument is the fourth one in `NewWriter`, and it should be `2`. The last cell on each line is not padded, so lines don't end in spaces.
- **`%!d(string=...)` in your output**: the verb doesn't match the value. Use `%d` for ints and `%s` for strings.
- **Still stuck?** Open `solution/receipt.go`, read it once, close it, then write your own version from memory.
