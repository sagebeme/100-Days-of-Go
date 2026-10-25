package pesabudget

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

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
	fs := flag.NewFlagSet("pesabudget", flag.ContinueOnError)
	fs.SetOutput(stderr)
	month := fs.String("month", "", "only this month, like 2026-04 (env PESA_MONTH)")
	budgetsFlag := fs.String("budgets", "", "budgets in shillings, like food=8000,transport=3500 (env PESA_BUDGETS)")
	strict := fs.Bool("strict", false, "exit with status 3 if a category is over budget (env PESA_STRICT)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	env := func(key string) (string, bool) {
		v, ok := lookup(key)
		return v, ok && v != ""
	}
	fail := func(code int, format string, a ...any) int {
		fmt.Fprintf(stderr, "pesabudget: "+format+"\n", a...)
		return code
	}

	if v, ok := env("PESA_MONTH"); ok && !set["month"] {
		*month = v
	}
	if *month != "" && !validMonth(*month) {
		return fail(ExitUsage, "month %q must look like 2026-04", *month)
	}

	budgetSource := "-budgets"
	if v, ok := env("PESA_BUDGETS"); ok && !set["budgets"] {
		*budgetsFlag, budgetSource = v, "PESA_BUDGETS"
	}
	budgets, err := ParseBudgets(*budgetsFlag)
	if err != nil {
		return fail(ExitUsage, "%s: %v", budgetSource, err)
	}

	if v, ok := env("PESA_STRICT"); ok && !set["strict"] {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fail(ExitUsage, "PESA_STRICT must be true or false, got %q", v)
		}
		*strict = b
	}

	extra := []Rule{}
	if v, ok := env("PESA_RULES"); ok {
		if extra, err = ParseRules(v); err != nil {
			return fail(ExitUsage, "PESA_RULES: %v", err)
		}
	}
	rules := append(extra, DefaultRules...)

	if fs.NArg() == 0 {
		return fail(ExitUsage, "give at least one statement CSV file")
	}
	var all []Transaction
	for _, path := range fs.Args() {
		txs, err := readStatement(path)
		if err != nil {
			return fail(ExitError, "%s: %v", path, err)
		}
		all = append(all, txs...)
	}

	s := Summarise(all, rules, budgets, *month)
	fmt.Fprint(stdout, Report(s))
	if *strict && s.Over() {
		return ExitOverBudget
	}
	return ExitOK
}

func readStatement(path string) ([]Transaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseStatement(f)
}
