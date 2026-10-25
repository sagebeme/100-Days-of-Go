package pesabudget

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// EAT is East Africa Time, UTC+3.
var EAT = time.FixedZone("EAT", 3*60*60)

// Transaction is one row of a statement. Amounts are in cents.
type Transaction struct {
	Receipt   string
	Time      time.Time
	Details   string
	Status    string
	PaidIn    int
	Withdrawn int // positive
	Balance   int
}

var columns = []string{"Receipt No.", "Completion Time", "Details", "Transaction Status", "Paid In", "Withdrawn", "Balance"}

// ParseStatement reads a statement CSV with the same rules as Day 24's Parse: columns
// found by name in any order, a byte order mark ignored, amounts in cents, Withdrawn
// stored as a positive number, and errors that start with "line N:".
func ParseStatement(r io.Reader) ([]Transaction, error) {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return nil, errors.New("line 1: empty statement, no header")
	}
	if err != nil {
		return nil, err
	}
	col := map[string]int{}
	for i, name := range header {
		if i == 0 {
			name = strings.TrimPrefix(name, "\ufeff")
		}
		col[strings.TrimSpace(name)] = i
	}
	for _, name := range columns {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("line 1: missing column %q", name)
		}
	}

	txs := []Transaction{}
	for {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return txs, nil
		}
		if err != nil {
			return nil, err
		}
		line, _ := cr.FieldPos(0)
		cell := func(name string) string { return strings.TrimSpace(row[col[name]]) }

		tx := Transaction{Receipt: cell("Receipt No."), Details: cell("Details"), Status: cell("Transaction Status")}
		if tx.Time, err = time.ParseInLocation("2006-01-02 15:04:05", cell("Completion Time"), EAT); err != nil {
			return nil, fmt.Errorf("line %d: Completion Time %q: want YYYY-MM-DD HH:MM:SS", line, cell("Completion Time"))
		}
		var out int
		for _, f := range []struct {
			name string
			dst  *int
		}{{"Paid In", &tx.PaidIn}, {"Withdrawn", &out}, {"Balance", &tx.Balance}} {
			if *f.dst, err = cents(cell(f.name)); err != nil {
				return nil, fmt.Errorf("line %d: %s: %w", line, f.name, err)
			}
		}
		if tx.PaidIn < 0 {
			return nil, fmt.Errorf("line %d: Paid In can't be negative", line)
		}
		if out > 0 {
			return nil, fmt.Errorf("line %d: Withdrawn should be negative", line)
		}
		tx.Withdrawn = -out
		txs = append(txs, tx)
	}
}

// cents reads "1,235.50", "-500.00", "750" or "" (0) as cents.
func cents(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	in := s
	sign := 1
	if strings.HasPrefix(s, "-") {
		sign, s = -1, s[1:]
	}
	whole, frac, hasDot := strings.Cut(strings.ReplaceAll(s, ",", ""), ".")
	if whole == "" || !digits(whole) || len(whole) > 12 || (hasDot && (len(frac) < 1 || len(frac) > 2 || !digits(frac))) {
		return 0, fmt.Errorf("%q isn't an amount", in)
	}
	for len(frac) < 2 {
		frac += "0"
	}
	n := 0
	for _, c := range whole + frac {
		n = n*10 + int(c-'0')
	}
	return sign * n, nil
}

func digits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
