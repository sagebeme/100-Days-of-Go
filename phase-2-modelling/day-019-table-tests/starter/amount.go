// Package tabletests parses and formats shilling amounts like "Ksh1,250.50". The
// code is small; the point of the day is the tests in amount_test.go.
package tabletests

import "errors"

// The ways ParseAmount can fail. Check them with errors.Is.
var (
	ErrEmpty    = errors.New("empty amount")
	ErrFormat   = errors.New("not an amount")
	ErrTooLarge = errors.New("amount too large")
)

// ParseAmount reads an amount of shillings and returns it in cents (hundredths of a
// shilling). It accepts:
//
//   - spaces around the amount, which it ignores
//   - an optional "Ksh" or "KES" prefix in any case, with optional spaces after it
//   - whole shillings, either plain ("1250") or with commas between groups of three
//     ("1,250", "12,345,678"); the first group has 1 to 3 digits
//   - optionally a "." and one or two digits of cents ("1250.5" is 125050)
//
// It returns ErrEmpty for "" or a lone prefix, ErrTooLarge for more than 9 digits of
// whole shillings, and ErrFormat for anything else it can't read. Errors are wrapped
// with the input, so check them with errors.Is.
func ParseAmount(s string) (int, error) {
	// Your code here. Let the failing subtests guide you, one at a time.
	return 0, nil
}

// FormatAmount writes cents as shillings with commas and two decimal places:
//
//	FormatAmount(125050)  → "Ksh1,250.50"
//	FormatAmount(5)       → "Ksh0.05"
//	FormatAmount(-20000)  → "-Ksh200.00"
func FormatAmount(cents int) string {
	// Your code here.
	return ""
}
