// Package invoice builds a one-page PDF invoice for a Kenyan business by hand: PDF
// objects, a cross-reference table and a trailer, with nothing but the standard library.
package invoice

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
func (l Line) Amount() int64 { return int64(l.Qty) * l.UnitCents }

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
	var t Totals
	for _, l := range lines {
		t.Subtotal += l.Amount()
	}
	t.VAT = (t.Subtotal*VATPercent + 50) / 100
	t.Total = t.Subtotal + t.VAT
	return t
}

// ValidKRAPIN reports whether s looks like a KRA PIN: 11 characters, a letter, nine
// digits and a letter, such as P051234567Q. Individuals' PINs start with A and other
// taxpayers' with P. Only the format is checked: whether the PIN is real is KRA's job.
func ValidKRAPIN(s string) bool {
	if len(s) != 11 || (s[0] != 'A' && s[0] != 'P') {
		return false
	}
	for i := 1; i <= 9; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s[10] >= 'A' && s[10] <= 'Z'
}

// FormatKsh writes cents as shillings with thousands separators: 116000 is
// "Ksh 1,160.00" and -5 is "-Ksh 0.05".
func FormatKsh(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	whole := strconv.FormatInt(cents/100, 10)
	var b strings.Builder
	for i, d := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}
	return fmt.Sprintf("%sKsh %s.%02d", sign, b.String(), cents%100)
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

// maxLineCents keeps every sum far from int64 overflow: Ksh 10 billion a line.
const maxLineCents = 1_000_000_000_000

// Validate checks the invoice before it's rendered, returning the first problem in
// the order the errors are listed above. A line over Ksh 10 billion (10^12 cents), or
// one whose Qty times UnitCents overflows, is ErrTooLarge.
func (inv Invoice) Validate() error {
	switch {
	case strings.TrimSpace(inv.Number) == "":
		return ErrNoNumber
	case inv.Date.IsZero():
		return ErrNoDate
	case strings.TrimSpace(inv.Seller.Name) == "":
		return ErrSellerName
	case !ValidKRAPIN(inv.Seller.KRAPIN):
		return ErrSellerPIN
	case inv.Buyer.KRAPIN != "" && !ValidKRAPIN(inv.Buyer.KRAPIN):
		return ErrBuyerPIN
	case len(inv.Lines) == 0:
		return ErrNoLines
	case len(inv.Lines) > MaxLines:
		return ErrTooManyLines
	}
	for _, l := range inv.Lines {
		if strings.TrimSpace(l.Description) == "" || l.Qty <= 0 || l.UnitCents < 0 {
			return ErrBadLine
		}
		if l.UnitCents > maxLineCents || l.Amount()/int64(l.Qty) != l.UnitCents || l.Amount() > maxLineCents {
			return ErrTooLarge
		}
	}
	return nil
}
