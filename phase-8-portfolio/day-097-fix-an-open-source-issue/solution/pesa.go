// Package pesa parses, formats and splits Kenyan shilling amounts without floating
// point. Amounts are held as whole cents in an int64, so 0.1 + 0.2 is exactly 0.3.
//
//	c, _ := pesa.Parse("Ksh 1,234.50") // 123450
//	pesa.Format(c)                     // "Ksh 1,234.50"
//	pesa.Split(c, 3)                   // [41150 41150 41150]
package pesa

import (
	"errors"
	"strconv"
	"strings"
)

// Cents is an amount of money in cents (1/100 of a shilling).
type Cents int64

// MaxDigits is the most digits Parse accepts before the decimal point. Thirteen
// digits of shillings is a trillion, and fits in an int64 of cents with room to spare.
const MaxDigits = 13

// ErrSyntax is returned (wrapped) by Parse for anything that isn't an amount.
var ErrSyntax = errors.New("pesa: invalid amount")

// Parse reads an amount such as "Ksh 1,234.50", "KES 99", "-ksh 20.5" or "1234".
//
// The grammar, after trimming spaces at both ends:
//
//	[-] [Ksh|KES] [spaces] [-] digits [. one or two digits]
//
// The currency is any case. The minus sign may come before or after the currency, but
// only once. Digits are either plain ("1234") or grouped in threes with commas
// ("1,234,567"). There must be at least one digit before the point.
func Parse(s string) (Cents, error) {
	in := s
	s = strings.TrimSpace(s)
	neg := false
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		neg, s = true, rest
	}
	if len(s) >= 3 && (strings.EqualFold(s[:3], "ksh") || strings.EqualFold(s[:3], "kes")) {
		s = strings.TrimLeft(s[3:], " ")
	}
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		if neg {
			return 0, syntaxError(in)
		}
		neg, s = true, rest
	}

	whole, frac, hasPoint := strings.Cut(s, ".")
	digits, ok := ungroup(whole)
	if !ok || len(digits) > MaxDigits {
		return 0, syntaxError(in)
	}
	if hasPoint && (len(frac) < 1 || len(frac) > 2 || !allDigits(frac)) {
		return 0, syntaxError(in)
	}

	shillings, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, syntaxError(in)
	}
	var cents int64
	if hasPoint {
		cents, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, syntaxError(in)
		}
		if len(frac) == 1 {
			cents *= 10 // "0.5" is 50 cents, not 5 (issue #42)
		}
	}

	total := Cents(shillings*100 + cents)
	if neg {
		total = -total
	}
	return total, nil
}

// ungroup checks the digits before the point and returns them without commas.
func ungroup(s string) (string, bool) {
	if !strings.Contains(s, ",") {
		return s, s != "" && allDigits(s)
	}
	groups := strings.Split(s, ",")
	if len(groups[0]) < 1 || len(groups[0]) > 3 || !allDigits(groups[0]) {
		return "", false
	}
	for _, g := range groups[1:] {
		if len(g) != 3 || !allDigits(g) {
			return "", false
		}
	}
	return strings.Join(groups, ""), true
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func syntaxError(s string) error {
	return &SyntaxError{Input: s}
}

// SyntaxError reports the input Parse couldn't read. errors.Is(err, ErrSyntax) is true.
type SyntaxError struct {
	Input string
}

func (e *SyntaxError) Error() string { return "pesa: invalid amount " + strconv.Quote(e.Input) }

// Unwrap makes errors.Is(err, ErrSyntax) work.
func (e *SyntaxError) Unwrap() error { return ErrSyntax }

// Format writes c as "Ksh 1,234.50", with "-Ksh" for a negative amount.
func Format(c Cents) string {
	sign := ""
	if c < 0 {
		sign = "-"
		c = -c
	}
	shillings := strconv.FormatInt(int64(c/100), 10)
	var b strings.Builder
	for i, r := range shillings {
		if i > 0 && (len(shillings)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	cents := int64(c % 100)
	return sign + "Ksh " + b.String() + "." + strconv.FormatInt(cents/10, 10) + strconv.FormatInt(cents%10, 10)
}

// Split shares c between n people as evenly as possible. The first c mod n people get
// one cent more, so the shares always add up to c. It panics if n < 1.
func Split(c Cents, n int) []Cents {
	if n < 1 {
		panic("pesa: Split into fewer than one share")
	}
	shares := make([]Cents, n)
	base, extra := c/Cents(n), c%Cents(n)
	step := Cents(1)
	if extra < 0 {
		extra, step = -extra, -1
	}
	for i := range shares {
		shares[i] = base
		if Cents(i) < extra {
			shares[i] += step
		}
	}
	return shares
}

// Sum adds amounts.
func Sum(amounts ...Cents) Cents {
	var total Cents
	for _, a := range amounts {
		total += a
	}
	return total
}
