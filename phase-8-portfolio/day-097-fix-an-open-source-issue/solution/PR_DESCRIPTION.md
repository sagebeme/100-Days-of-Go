## What this fixes

`Parse` read one digit after the point as cents, so `Parse("Ksh 1,200.5")` returned 120005 (Ksh 1,200.05) instead of 120050 (Ksh 1,200.50).

Fixes #42

## The cause

In `Parse`, the digits after the point go through `strconv.ParseInt` and are added as cents. That's right for two digits (`"50"` is 50 cents), but `"5"` also parses as 5, while it means 5 tenths of a shilling, which is 50 cents. The existing tests only used two-digit fractions, so they didn't catch it.

## The fix

When the fraction has one digit, multiply it by 10 before adding it. That's a three-line change in `Parse`. The grammar, the doc comment and the public API don't change: one or two digits after the point were already allowed.

I chose this over padding the string to two digits (`frac + "0"`) because it keeps the parsing in one place and allocates nothing.

## How I tested it

- Added `TestParseOneFractionDigit` with the input from the issue, a negative amount and a two-digit amount that must not change. It failed before the fix with `Parse("Ksh 1,200.5") = 120005` and passes after.
- `gofmt -l .` prints nothing, and `go vet ./...` and `go test -race ./...` pass.

## Out of scope

Amounts with more than two decimal places (`"1.505"`) are still rejected, as before. Rounding them would be a behaviour change, so it needs its own issue if anyone wants it.
