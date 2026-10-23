// Package receiptprinter formats shilling amounts and prints a duka receipt in neat columns.
package receiptprinter

// Commas writes n with a comma between each group of three digits.
//
//	Commas(1250000) → "1,250,000"
func Commas(n int) string {
	// Your code here. Follow the steps in the README.
	return ""
}

// Shillings formats an amount given in cents as shillings with commas and two decimal
// places. A negative amount puts the minus sign before "Ksh".
//
//	Shillings(125000) → "Ksh1,250.00"
func Shillings(cents int) string {
	// Your code here.
	return ""
}

// Line formats one row of a receipt in fixed columns: the name left-aligned in 16
// characters (cut short if it is longer), the quantity right-aligned in 3, and the row
// total right-aligned in 14, with one space between columns.
//
//	Line("Unga 2kg", 2, 23000) → "Unga 2kg           2      Ksh460.00"
func Line(name string, qty, priceCents int) string {
	// Your code here.
	return ""
}

// Receipt lays out a whole receipt with text/tabwriter: a header row, one row per item
// and a TOTAL row, in columns separated by at least two spaces. items, qtys and
// priceCents must be the same length, every quantity must be at least 1 and no price
// can be negative.
func Receipt(items []string, qtys []int, priceCents []int) (string, error) {
	// Your code here.
	return "", nil
}
