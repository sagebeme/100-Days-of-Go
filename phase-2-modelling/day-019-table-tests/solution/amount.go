// Package tabletests parses and formats shilling amounts like "Ksh1,250.50". The
// code is small; the point of the day is the tests in amount_test.go.
package tabletests

import (
	"errors"
	"fmt"
	"strings"
)

// The ways ParseAmount can fail. Check them with errors.Is.
var (
	ErrEmpty    = errors.New("empty amount")
	ErrFormat   = errors.New("not an amount")
	ErrTooLarge = errors.New("amount too large")
)

// maxDigits is the most digits allowed before the decimal point.
const maxDigits = 9

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
	in := s
	s = strings.TrimSpace(s)
	if len(s) >= 3 {
		if p := strings.ToLower(s[:3]); p == "ksh" || p == "kes" {
			s = strings.TrimSpace(s[3:])
		}
	}
	if s == "" {
		return 0, fmt.Errorf("%w: %q", ErrEmpty, in)
	}

	whole, frac, hasDot := strings.Cut(s, ".")
	digits, ok := wholeDigits(whole)
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrFormat, in)
	}
	if hasDot && (len(frac) < 1 || len(frac) > 2 || !allDigits(frac)) {
		return 0, fmt.Errorf("%w: %q", ErrFormat, in)
	}
	if len(digits) > maxDigits {
		return 0, fmt.Errorf("%w: %q", ErrTooLarge, in)
	}

	cents := 0
	for i := 0; i < len(digits); i++ {
		cents = cents*10 + int(digits[i]-'0')
	}
	cents *= 100
	if len(frac) == 1 {
		frac += "0"
	}
	if hasDot {
		cents += int(frac[0]-'0')*10 + int(frac[1]-'0')
	}
	return cents, nil
}

// wholeDigits checks the part before the decimal point and returns its digits without
// commas.
func wholeDigits(s string) (string, bool) {
	if !strings.Contains(s, ",") {
		return s, s != "" && allDigits(s)
	}
	groups := strings.Split(s, ",")
	if len(groups[0]) < 1 || len(groups[0]) > 3 {
		return "", false
	}
	for _, g := range groups[1:] {
		if len(g) != 3 {
			return "", false
		}
	}
	digits := strings.Join(groups, "")
	return digits, allDigits(digits)
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// FormatAmount writes cents as shillings with commas and two decimal places:
//
//	FormatAmount(125050)  → "Ksh1,250.50"
//	FormatAmount(5)       → "Ksh0.05"
//	FormatAmount(-20000)  → "-Ksh200.00"
func FormatAmount(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	whole := fmt.Sprint(cents / 100)
	var b strings.Builder
	for i := 0; i < len(whole); i++ {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(whole[i])
	}
	return fmt.Sprintf("%sKsh%s.%02d", sign, b.String(), cents%100)
}
