package pesabudget

import (
	"io"
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
	// Your code here.
	return nil, nil
}
