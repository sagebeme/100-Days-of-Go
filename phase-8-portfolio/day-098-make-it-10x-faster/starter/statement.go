// Package statement totals an M-Pesa-style statement by counterparty.
//
// This version works, and every correctness test passes. It is also very slow. Your
// job is to make it at least 10 times faster without changing what it returns.
package statement

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
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
func Summarise(statement string) ([]Total, error) {
	totals := []Total{}
	lines := strings.Split(statement, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || line == "receipt,date,counterparty,amount" {
			continue
		}
		re := regexp.MustCompile(`^([A-Z0-9]{10}),([^,]+),([^,]+),(-?[0-9]{1,12}\.[0-9]{2})$`)
		m := re.FindStringSubmatch(line)
		if m == nil {
			return nil, &LineError{Line: i + 1, Reason: fmt.Sprintf("can't read %q", line)}
		}
		name := strings.ToUpper(strings.TrimSpace(m[3]))
		if name == "" {
			return nil, &LineError{Line: i + 1, Reason: "blank counterparty"}
		}
		amount, err := strconv.ParseFloat(m[4], 64)
		if err != nil {
			return nil, &LineError{Line: i + 1, Reason: err.Error()}
		}
		cents := int64(math.Round(amount * 100))

		found := -1
		for j := range totals {
			if fmt.Sprintf("%s", totals[j].Name) == name {
				found = j
			}
		}
		if found == -1 {
			totals = append(totals, Total{Name: name})
			found = len(totals) - 1
		}
		t := totals[found]
		t.Count++
		if cents < 0 {
			t.OutCents += -cents
		} else {
			t.InCents += cents
		}
		totals[found] = t

		// Keep the result sorted as we go, so it's ready at the end.
		sorted := make([]Total, len(totals))
		copy(sorted, totals)
		sort.Slice(sorted, func(a, b int) bool {
			if sorted[a].OutCents != sorted[b].OutCents {
				return sorted[a].OutCents > sorted[b].OutCents
			}
			if sorted[a].InCents != sorted[b].InCents {
				return sorted[a].InCents > sorted[b].InCents
			}
			return sorted[a].Name < sorted[b].Name
		})
		totals = sorted
	}
	return totals, nil
}
