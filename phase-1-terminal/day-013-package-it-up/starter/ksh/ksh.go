// Package ksh formats whole-shilling amounts the way Kenyan shops print them.
package ksh

// Format writes amount in shillings with "Ksh" in front and a comma between each
// group of three digits. A negative amount puts the minus sign before "Ksh".
//
//	Format(1250) → "Ksh1,250"
func Format(amount int) string {
	// Your code here. Put the comma logic in a small unexported helper, group(n int) string,
	// and call it from here.
	return ""
}
