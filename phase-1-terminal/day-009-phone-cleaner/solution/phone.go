// Package phonecleaner turns Kenyan mobile numbers, typed any which way, into one standard form.
package phonecleaner

import (
	"errors"
	"fmt"
	"strings"
)

// Clean returns raw in the form 2547XXXXXXXX or 2541XXXXXXXX. It accepts numbers
// written as 0712 345 678, 0712-345-678, 712345678, 254712345678 or +254712345678.
// Spaces and hyphens are ignored. A plus sign is allowed only at the very start.
// Anything else gives an error.
//
//	Clean("0712 345 678") → "254712345678", nil
func Clean(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty phone number")
	}

	plus := false
	digits := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits += string(c)
		case c == ' ' || c == '-':
			// separators people type: skip them
		case c == '+' && i == 0:
			plus = true
		default:
			return "", fmt.Errorf("invalid character %q in %q", c, raw)
		}
	}

	var local string
	switch {
	case len(digits) == 12 && strings.HasPrefix(digits, "254"):
		local = digits[3:]
	case plus:
		return "", fmt.Errorf("%q: a number starting with + must start with +254", raw)
	case len(digits) == 10 && digits[0] == '0':
		local = digits[1:]
	case len(digits) == 9:
		local = digits
	default:
		return "", fmt.Errorf("%q has %d digits, which isn't a Kenyan mobile number", raw, len(digits))
	}

	if local[0] != '7' && local[0] != '1' {
		return "", fmt.Errorf("%q isn't a mobile number: it should start 07 or 01", raw)
	}
	return "254" + local, nil
}

// Mask cleans raw and hides the middle digits, the way M-Pesa shows a number on a
// receipt.
//
//	Mask("+254712345678") → "0712***678", nil
func Mask(raw string) (string, error) {
	clean, err := Clean(raw)
	if err != nil {
		return "", err
	}
	local := clean[3:]
	return "0" + local[:3] + "***" + local[6:], nil
}

// SameNumber reports whether a and b are the same valid phone number written two
// different ways. An invalid number is never the same as anything.
//
//	SameNumber("0712 345 678", "+254712345678") → true
func SameNumber(a, b string) bool {
	ca, err := Clean(a)
	if err != nil {
		return false
	}
	cb, err := Clean(b)
	if err != nil {
		return false
	}
	return ca == cb
}
