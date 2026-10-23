// Package receiptprinter formats shilling amounts and prints a duka receipt in neat columns.
package receiptprinter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Commas writes n with a comma between each group of three digits.
//
//	Commas(1250000) → "1,250,000"
func Commas(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := strconv.Itoa(n)
	var b strings.Builder
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(digits[i])
	}
	return sign + b.String()
}

// Shillings formats an amount given in cents as shillings with commas and two decimal
// places. A negative amount puts the minus sign before "Ksh".
//
//	Shillings(125000) → "Ksh1,250.00"
func Shillings(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%sKsh%s.%02d", sign, Commas(cents/100), cents%100)
}

// Line formats one row of a receipt in fixed columns: the name left-aligned in 16
// characters (cut short if it is longer), the quantity right-aligned in 3, and the row
// total right-aligned in 14, with one space between columns.
//
//	Line("Unga 2kg", 2, 23000) → "Unga 2kg           2      Ksh460.00"
func Line(name string, qty, priceCents int) string {
	return fmt.Sprintf("%-16.16s %3d %14s", name, qty, Shillings(qty*priceCents))
}

// Receipt lays out a whole receipt with text/tabwriter: a header row, one row per item
// and a TOTAL row, in columns separated by at least two spaces. items, qtys and
// priceCents must be the same length, every quantity must be at least 1 and no price
// can be negative.
func Receipt(items []string, qtys []int, priceCents []int) (string, error) {
	if len(items) != len(qtys) || len(items) != len(priceCents) {
		return "", fmt.Errorf("got %d items, %d quantities and %d prices; they must match", len(items), len(qtys), len(priceCents))
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Item\tQty\tTotal")
	total := 0
	for i, item := range items {
		if qtys[i] < 1 {
			return "", fmt.Errorf("%s: quantity %d must be at least 1", item, qtys[i])
		}
		if priceCents[i] < 0 {
			return "", errors.New(item + ": price can't be negative")
		}
		row := qtys[i] * priceCents[i]
		total += row
		fmt.Fprintf(w, "%s\t%d\t%s\n", item, qtys[i], Shillings(row))
	}
	fmt.Fprintf(w, "TOTAL\t\t%s\n", Shillings(total))
	if err := w.Flush(); err != nil {
		return "", err
	}
	return b.String(), nil
}
