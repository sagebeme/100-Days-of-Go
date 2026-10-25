// Package invoice builds a one-page PDF invoice for a Kenyan business by hand: PDF
// objects, a cross-reference table and a trailer, with nothing but the standard library.
package invoice

import (
	"errors"
	"fmt"
	"time"
)

// VATPercent is Kenya's standard VAT rate.
const VATPercent = 16

// MaxLines is how many lines fit on the one page.
const MaxLines = 20

// Party is a seller or a buyer.
type Party struct {
	Name    string
	KRAPIN  string // required for the seller, optional for the buyer
	Address string
	Phone   string
}

// Line is one row of the invoice. Prices are in cents (1 Ksh = 100 cents) and
// exclude VAT.
type Line struct {
	Description string
	Qty         int
	UnitCents   int64
}

// Amount is Qty times UnitCents.
func (l Line) Amount() int64 {
	// Your code here
	return 0
}

// Invoice is everything printed on the page.
type Invoice struct {
	Number string
	Date   time.Time
	Seller Party
	Buyer  Party
	Lines  []Line
}

// Totals are in cents.
type Totals struct {
	Subtotal int64
	VAT      int64
	Total    int64
}

// ComputeTotals adds up the lines and charges VATPercent on the subtotal, rounded to
// the nearest cent. (16% of a whole number of cents is never exactly half a cent, so
// there are no ties to break.)
func ComputeTotals(lines []Line) Totals {
	// Your code here
	return Totals{}
}

// ValidKRAPIN reports whether s looks like a KRA PIN: 11 characters, a letter, nine
// digits and a letter, such as P051234567Q. Individuals' PINs start with A and other
// taxpayers' with P. Only the format is checked: whether the PIN is real is KRA's job.
func ValidKRAPIN(s string) bool {
	// Your code here
	return false
}

// FormatKsh writes cents as shillings with thousands separators: 116000 is
// "Ksh 1,160.00" and -5 is "-Ksh 0.05".
func FormatKsh(cents int64) string {
	// Your code here
	return ""
}

// Validation errors.
var (
	ErrNoNumber     = errors.New("invoice number is required")
	ErrNoDate       = errors.New("invoice date is required")
	ErrSellerName   = errors.New("seller name is required")
	ErrSellerPIN    = errors.New("seller KRA PIN must look like P051234567Q")
	ErrBuyerPIN     = errors.New("buyer KRA PIN, when given, must look like A123456789B")
	ErrNoLines      = errors.New("an invoice needs at least one line")
	ErrTooManyLines = fmt.Errorf("at most %d lines fit on one page", MaxLines)
	ErrBadLine      = errors.New("each line needs a description, a quantity above 0 and a price of 0 or more")
	ErrTooLarge     = errors.New("amounts are too large")
)

// Validate checks the invoice before it's rendered, returning the first problem in
// the order the errors are listed above. A line over Ksh 10 billion (10^12 cents), or
// one whose Qty times UnitCents overflows, is ErrTooLarge.
func (inv Invoice) Validate() error {
	// Your code here
	return nil
}
