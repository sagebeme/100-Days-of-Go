// Package pesabudget is the Phase 2 capstone: a command-line budget tracker that reads
// M-Pesa statement CSVs, sorts the spending into categories, compares it with budgets
// and prints a report. It's configured with env variables and flags.
package pesabudget

// Rule puts a transaction in Category when Keyword appears in its Details, ignoring
// case.
type Rule struct {
	Keyword  string
	Category string
}

// DefaultRules are tried after any rules from PESA_RULES. Order matters: the first
// rule that matches wins, which is why "Charge" comes first.
var DefaultRules = []Rule{
	{"Charge", "fees"},
	{"SOKO BORA", "food"},
	{"MAMA PIMA", "food"},
	{"NYAMA CHOMA", "food"},
	{"STIMA", "bills"},
	{"MAJI", "bills"},
	{"SACCO", "transport"},
	{"Airtime", "airtime"},
	{"Customer Transfer", "sent"},
}

// Other is the category for a transaction no rule matches.
const Other = "other"

// Categorise returns the Category of the first rule whose Keyword appears in details,
// ignoring case, or Other.
//
//	Categorise("Merchant Payment to 512345 - SOKO BORA SUPERMARKET", DefaultRules) → "food"
func Categorise(details string, rules []Rule) string {
	// Your code here.
	return ""
}

// ParseRules reads rules in the form PESA_RULES uses: KEYWORD=category pairs separated
// by ";", with spaces around each part trimmed. The category is lower-cased. "" gives
// no rules. A pair with no "=" or an empty side is an error.
//
//	ParseRules("JAVA HOUSE=eating-out; Netflix=fun") → [{JAVA HOUSE eating-out} {Netflix fun}]
func ParseRules(s string) ([]Rule, error) {
	// Your code here.
	return nil, nil
}

// ParseBudgets reads budgets in the form PESA_BUDGETS and -budgets use:
// category=amount pairs separated by ",", amounts in whole shillings, spaces trimmed,
// categories lower-cased. It returns them in cents. "" gives an empty map.
//
//	ParseBudgets("food=8000, transport=3500") → map[food:800000 transport:350000]
//
// Errors:
//
//	bad budget "food"; want category=amount
//	budget for "food" must be a whole number of shillings, got "8k"
//	budget for "food" given twice
func ParseBudgets(s string) (map[string]int, error) {
	// Your code here.
	return nil, nil
}

// Line is one category in the report. Amounts are in cents.
type Line struct {
	Category  string
	Spent     int
	Budget    int
	HasBudget bool
}

// Summary is everything the report shows.
type Summary struct {
	Month  string // "2026-04", or "" for every month
	Lines  []Line
	PaidIn int // cents
}

// Summarise adds up the completed transactions in month ("2026-04", or "" for all) by
// category. Only withdrawals count as spending; PaidIn adds up the money received.
// A receipt number seen more than once (statements that overlap) counts only the
// first time. Every category with a budget gets a Line, even with nothing spent.
// Lines are sorted by Spent, highest first, then by Category.
func Summarise(txs []Transaction, rules []Rule, budgets map[string]int, month string) Summary {
	// Your code here.
	return Summary{}
}

// Over reports whether any category has spent more than its budget.
func (s Summary) Over() bool {
	// Your code here.
	return false
}

// Report formats a Summary as a table. Each row is fmt.Sprintf("%-12s %14s %14s %14s")
// of the category, spent, budget and left (budget minus spent), with "-" for the
// budget and left of a category that has no budget, and " OVER" after a row that is
// over its budget. The TOTAL row adds up spending, budgets, and left over the
// budgeted categories; its budget and left are "-" when there are no budgets.
//
//	Pesa Budget: 2026-04
//	CATEGORY              SPENT         BUDGET           LEFT
//	food            Ksh4,100.00    Ksh3,000.00   -Ksh1,100.00 OVER
//	bills           Ksh1,600.00              -              -
//	TOTAL           Ksh5,700.00    Ksh3,000.00   -Ksh1,100.00
//	money in: Ksh8,000.00
//
// The title says "all months" when Month is "". Every line ends with "\n".
func Report(s Summary) string {
	// Your code here.
	return ""
}

// Ksh formats cents as shillings with commas and two decimal places:
// Ksh(410000) is "Ksh4,100.00" and Ksh(-110000) is "-Ksh1,100.00".
func Ksh(cents int) string {
	// Your code here.
	return ""
}
