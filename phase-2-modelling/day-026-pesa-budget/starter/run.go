package pesabudget

import "io"

// The exit codes Run returns.
const (
	ExitOK         = 0
	ExitError      = 1 // a statement file couldn't be read or parsed
	ExitUsage      = 2 // bad flags, env variables or arguments
	ExitOverBudget = 3 // -strict, and a category is over budget
)

// LookupFunc finds an env variable, like os.LookupEnv.
type LookupFunc func(key string) (string, bool)

// Run is the pesabudget command:
//
//	pesabudget [-month 2026-04] [-budgets food=8000,transport=3500] [-strict] statement.csv...
//
// Settings come from a flag if it's given, otherwise from an env variable read with
// lookup, otherwise a default. An env variable set to "" counts as not set.
//
//	-month    PESA_MONTH    only this month, YYYY-MM; default every month
//	-budgets  PESA_BUDGETS  budgets, as ParseBudgets reads them; default none
//	-strict   PESA_STRICT   exit with ExitOverBudget if any category is over; default false
//	          PESA_RULES    extra rules, as ParseRules reads them, tried before DefaultRules
//
// It reads every file named after the flags (at least one), summarises them together
// and prints the Report to stdout. Problems go to stderr as one line starting with
// "pesabudget: ". A bad flag, setting or missing file name returns ExitUsage; a file
// that can't be opened or parsed returns ExitError and its message starts with the file
// name: "pesabudget: april.csv: line 3: ...".
func Run(args []string, lookup LookupFunc, stdout, stderr io.Writer) int {
	// Your code here.
	return 0
}
