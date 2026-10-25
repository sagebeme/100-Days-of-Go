// Package csvstatements reads an M-Pesa statement exported as CSV. The format is
// modelled on the columns of a real statement; the data in testdata/ is made up.
package csvstatements

import (
	"io"
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
	// Your code here.
	return nil, nil
}

// Totals adds up the money in and out over the completed transactions only. A failed
// or reversed payment never moved any money.
func Totals(txs []Transaction) (in, out int) {
	// Your code here.
	return 0, 0
}

// Counterparty returns who the money went to or came from: the text after the first
// " - " in details, with spaces trimmed. Details with no " - " (like "Airtime
// Purchase") are returned whole.
//
//	Counterparty("Merchant Payment to 512345 - SOKO BORA SUPERMARKET") → "SOKO BORA SUPERMARKET"
func Counterparty(details string) string {
	// Your code here.
	return ""
}

// SpentWith adds up the completed withdrawals for each counterparty, in cents.
// Counterparties you only received money from don't appear.
func SpentWith(txs []Transaction) map[string]int {
	// Your code here.
	return nil
}
