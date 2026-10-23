// Package ksh formats whole-shilling amounts the way Kenyan shops print them.
package ksh

import "strconv"

// Format writes amount in shillings with "Ksh" in front and a comma between each
// group of three digits. A negative amount puts the minus sign before "Ksh".
//
//	Format(1250) → "Ksh1,250"
func Format(amount int) string {
	if amount < 0 {
		return "-Ksh" + group(-amount)
	}
	return "Ksh" + group(amount)
}

// group puts commas into a non-negative number: 1250000 becomes "1,250,000".
// Its name starts with a small letter, so code outside package ksh can't call it.
func group(n int) string {
	digits := strconv.Itoa(n)
	out := ""
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out += ","
		}
		out += string(digits[i])
	}
	return out
}
