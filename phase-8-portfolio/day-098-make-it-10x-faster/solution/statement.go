// Package statement totals an M-Pesa-style statement by counterparty.
package statement

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// Total is what one counterparty adds up to across the statement.
type Total struct {
	Name     string // upper case, spaces at the ends trimmed
	Count    int    // number of transactions
	InCents  int64  // money received from them (positive amounts)
	OutCents int64  // money paid to them (negative amounts), as a positive number
}

// LineError reports a line Summarise couldn't read. Line counts from 1, header included.
type LineError struct {
	Line   int
	Reason string
}

func (e *LineError) Error() string { return fmt.Sprintf("statement: line %d: %s", e.Line, e.Reason) }

const header = "receipt,date,counterparty,amount"

// Summarise reads a statement, one transaction per line:
//
//	receipt,date,counterparty,amount
//	SJK2X1ABCD,2026-10-01 08:15,MAMA MBOGA KAWANGWARE,-350.00
//
// The receipt is exactly 10 characters of A-Z and 0-9. The date is any non-empty text
// without a comma. The counterparty is any text without a comma that isn't blank once
// trimmed; names are matched ignoring case. The amount is an optional minus, 1 to 12
// digits, a point and exactly two digits. A line that is exactly the header
// "receipt,date,counterparty,amount", or blank, is skipped. Lines may end in \r\n.
//
// It returns one Total per counterparty, sorted by OutCents (largest first), then
// InCents (largest first), then Name. An empty statement gives an empty slice.
//
// It makes one pass over the text, slices it without copying, looks names up in a map,
// and sorts once at the end: O(n + k log k) for n lines and k counterparties.
func Summarise(statement string) ([]Total, error) {
	index := map[string]int{} // name -> position in totals
	totals := []Total{}
	rest := statement
	for n := 1; rest != ""; n++ {
		var line string
		line, rest, _ = strings.Cut(rest, "\n")
		line = strings.TrimSuffix(line, "\r")
		if line == header || strings.TrimSpace(line) == "" {
			continue
		}
		name, cents, reason := parseLine(line)
		if reason != "" {
			return nil, &LineError{Line: n, Reason: reason}
		}
		i, ok := index[name]
		if !ok {
			i = len(totals)
			index[name] = i
			totals = append(totals, Total{Name: name})
		}
		t := &totals[i]
		t.Count++
		if cents < 0 {
			t.OutCents -= cents
		} else {
			t.InCents += cents
		}
	}
	slices.SortFunc(totals, func(a, b Total) int {
		return cmp.Or(
			cmp.Compare(b.OutCents, a.OutCents),
			cmp.Compare(b.InCents, a.InCents),
			strings.Compare(a.Name, b.Name),
		)
	})
	return totals, nil
}

// parseLine splits one line into its four fields and checks them. It returns the
// upper-case name and the amount in cents, or a reason when the line is bad.
// strings.Cut returns substrings of line, so nothing is copied; strings.ToUpper only
// allocates when the name has a lower-case letter in it.
func parseLine(line string) (name string, cents int64, reason string) {
	receipt, rest, ok1 := strings.Cut(line, ",")
	date, rest, ok2 := strings.Cut(rest, ",")
	party, amount, ok3 := strings.Cut(rest, ",")
	if !ok1 || !ok2 || !ok3 || strings.Contains(amount, ",") {
		return "", 0, "want 4 comma-separated fields"
	}
	if !validReceipt(receipt) {
		return "", 0, "bad receipt " + receipt
	}
	if date == "" {
		return "", 0, "missing date"
	}
	name = strings.ToUpper(strings.TrimSpace(party))
	if name == "" {
		return "", 0, "blank counterparty"
	}
	cents, ok := parseCents(amount)
	if !ok {
		return "", 0, "bad amount " + amount
	}
	return name, cents, ""
}

func validReceipt(s string) bool {
	if len(s) != 10 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// parseCents reads -?d{1,12}.dd into cents without floating point.
func parseCents(s string) (int64, bool) {
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg, s = true, s[1:]
	}
	if len(s) < 4 || len(s) > 15 || s[len(s)-3] != '.' {
		return 0, false
	}
	var v int64
	for i := 0; i < len(s); i++ {
		if i == len(s)-3 {
			continue // the point
		}
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		v = v*10 + int64(c-'0')
	}
	if neg {
		v = -v
	}
	return v, true
}
