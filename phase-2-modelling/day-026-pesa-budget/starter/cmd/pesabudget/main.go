// Command pesabudget reports M-Pesa spending against budgets.
//
//	PESA_BUDGETS=food=3000,transport=500 go run ./phase-2-modelling/day-026-pesa-budget/starter/cmd/pesabudget \
//	    -month 2026-04 phase-2-modelling/day-026-pesa-budget/starter/testdata/april.csv
package main

import (
	"os"

	pesabudget "github.com/sagebeme/100-Days-of-Go/phase-2-modelling/day-026-pesa-budget/starter"
)

func main() {
	os.Exit(pesabudget.Run(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}
