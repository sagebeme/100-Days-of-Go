// Package cliflags is the chama command-line tool: it splits a contribution between
// members and works out loan repayments. Run is the whole program, written so tests
// can call it with any arguments and read what it prints.
package cliflags

import (
	"io"
	"strconv"
)

// Version is what "chama version" prints after "chama ".
const Version = "1.0.0"

// The exit codes Run returns.
const (
	ExitOK    = 0 // it worked
	ExitError = 1 // the input made sense but the command couldn't do it
	ExitUsage = 2 // the command line was wrong: unknown command, bad or missing flag
)

// Usage is printed for "chama help", and to stderr for a missing or unknown command.
const Usage = `usage: chama <command> [flags]

commands:
  split    share a total between members
  loan     work out a chama loan's repayments
  version  print the version

Run 'chama <command> -h' for a command's flags.
`

// Split shares total shillings between members as evenly as whole shillings allow.
// When it doesn't divide evenly, the first members pay one shilling more each:
//
//	Split(12005, 5) → [2401 2401 2401 2401 2401]
//	Split(10, 3)    → [4 3 3]
//
// It returns an error if members is below 1 or total is below members (somebody would
// pay nothing).
func Split(total, members int) ([]int, error) {
	// Your code here.
	return nil, nil
}

// Loan works out a flat-rate chama loan: interest is ratePercent of amount for every
// month, rounded down; total is amount plus interest; monthly is total divided by
// months, rounded up so the loan is fully repaid.
//
//	Loan(20000, 10, 3) → interest 6000, total 26000, monthly 8667
func Loan(amount, ratePercent, months int) (interest, total, monthly int) {
	// Your code here.
	return 0, 0, 0
}

// Run runs the chama tool with args (the command line without the program name),
// writes results to stdout and problems to stderr, and returns the exit code.
//
//	split -total N -members M   one line per member: "member 1: Ksh2,401"
//	loan -amount N [-rate R] [-months M]
//	                            "interest: Ksh6,000", "total: Ksh26,000", "monthly: Ksh8,667"
//	                            -rate is percent per month, 0 to 100, default 10;
//	                            -months is 1 to 36, default 3
//	version                     "chama 1.0.0"
//	help                        Usage, on stdout
//
// Flags -total, -members and -amount are required and must be at least 1. A problem
// with the command line prints "chama: <problem>" or "chama <command>: <problem>" on
// stderr and returns ExitUsage; so does a missing or unknown command, after printing
// Usage to stderr. Extra arguments after the flags are a usage error too. "-h" on a
// command prints its flags and returns ExitOK. A split that can't be done returns
// ExitError.
func Run(args []string, stdout, stderr io.Writer) int {
	// Your code here. Switch on args[0], and give each command its own
	// flag.NewFlagSet(name, flag.ContinueOnError) with SetOutput(stderr).
	return 0
}

// ksh formats a whole number of shillings with commas: ksh(26000) is "Ksh26,000".
func ksh(n int) string {
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}
	digits := strconv.Itoa(n)
	out := ""
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out += ","
		}
		out += string(digits[i])
	}
	return sign + "Ksh" + out
}
