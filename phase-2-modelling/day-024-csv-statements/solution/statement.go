// Package csvstatements reads an M-Pesa statement exported as CSV. The format is
// modelled on the columns of a real statement; the data in testdata/ is made up.
package csvstatements

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// EAT is East Africa Time, UTC+3, the time zone statement times are written in.
var EAT = time.FixedZone("EAT", 3*60*60)

// TimeLayout is how the Completion Time column writes a time.
const TimeLayout = "2006-01-02 15:04:05"

// The columns Parse needs. They may come in any order, and other columns are ignored.
var Columns = []string{"Receipt No.", "Completion Time", "Details", "Transaction Status", "Paid In", "Withdrawn", "Balance"}

// Transaction is one row of a statement. Amounts are in cents.
type Transaction struct {
	Receipt   string
	Time      time.Time // in EAT
	Details   string
	Status    string // "Completed", "Failed", ...
	PaidIn    int    // cents, 0 or more
	Withdrawn int    // cents, 0 or more (the file writes it with a minus sign)
	Balance   int    // cents, may be negative after a Fuliza overdraft
}

// Parse reads a statement: a header row, then one row per transaction.
//
// It finds the columns by their names in the header, so their order doesn't matter.
// It ignores spaces around header names and a UTF-8 byte order mark before the first
// one (spreadsheet programs add it). An empty amount cell means 0. Amounts may have
// commas between thousands and up to two decimal places: "1,235.50", "-500.00", "750".
// Withdrawn amounts in the file are negative or zero; Parse stores them as positive.
//
// It returns an error, starting with "line N:", for a missing column, a row with the
// wrong number of cells, a time that doesn't match TimeLayout, an amount it can't
// read, a negative Paid In or a positive Withdrawn. An empty input is an error too.
func Parse(r io.Reader) ([]Transaction, error) {
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
	for _, name := range Columns {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("line 1: missing column %q", name)
		}
	}

	txs := []Transaction{}
	for {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err // a *csv.ParseError already says "record on line N"
		}
		line, _ := cr.FieldPos(0)
		tx, err := parseRow(row, col)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func parseRow(row []string, col map[string]int) (Transaction, error) {
	cell := func(name string) string { return strings.TrimSpace(row[col[name]]) }

	tx := Transaction{
		Receipt: cell("Receipt No."),
		Details: cell("Details"),
		Status:  cell("Transaction Status"),
	}
	t, err := time.ParseInLocation(TimeLayout, cell("Completion Time"), EAT)
	if err != nil {
		return Transaction{}, fmt.Errorf("Completion Time %q: want YYYY-MM-DD HH:MM:SS", cell("Completion Time"))
	}
	tx.Time = t

	if tx.PaidIn, err = cents(cell("Paid In")); err != nil {
		return Transaction{}, fmt.Errorf("Paid In: %w", err)
	}
	if tx.PaidIn < 0 {
		return Transaction{}, fmt.Errorf("Paid In %q can't be negative", cell("Paid In"))
	}
	out, err := cents(cell("Withdrawn"))
	if err != nil {
		return Transaction{}, fmt.Errorf("Withdrawn: %w", err)
	}
	if out > 0 {
		return Transaction{}, fmt.Errorf("Withdrawn %q should be negative", cell("Withdrawn"))
	}
	tx.Withdrawn = -out
	if tx.Balance, err = cents(cell("Balance")); err != nil {
		return Transaction{}, fmt.Errorf("Balance: %w", err)
	}
	return tx, nil
}

// cents reads "1,235.50", "-500.00", "750" or "" (which is 0) as cents.
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
	if whole == "" || !digits(whole) || (hasDot && (len(frac) < 1 || len(frac) > 2 || !digits(frac))) {
		return 0, fmt.Errorf("%q isn't an amount", in)
	}
	if len(whole) > 12 {
		return 0, fmt.Errorf("%q is too large", in)
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

// Totals adds up the money in and out over the completed transactions only. A failed
// or reversed payment never moved any money.
func Totals(txs []Transaction) (in, out int) {
	for _, tx := range txs {
		if tx.Status != "Completed" {
			continue
		}
		in += tx.PaidIn
		out += tx.Withdrawn
	}
	return in, out
}

// Counterparty returns who the money went to or came from: the text after the first
// " - " in details, with spaces trimmed. Details with no " - " (like "Airtime
// Purchase") are returned whole.
//
//	Counterparty("Merchant Payment to 512345 - SOKO BORA SUPERMARKET") → "SOKO BORA SUPERMARKET"
func Counterparty(details string) string {
	if _, after, ok := strings.Cut(details, " - "); ok {
		return strings.TrimSpace(after)
	}
	return strings.TrimSpace(details)
}

// SpentWith adds up the completed withdrawals for each counterparty, in cents.
// Counterparties you only received money from don't appear.
func SpentWith(txs []Transaction) map[string]int {
	spent := map[string]int{}
	for _, tx := range txs {
		if tx.Status == "Completed" && tx.Withdrawn > 0 {
			spent[Counterparty(tx.Details)] += tx.Withdrawn
		}
	}
	return spent
}
