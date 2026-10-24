// Package cliflags is the chama command-line tool: it splits a contribution between
// members and works out loan repayments. Run is the whole program, written so tests
// can call it with any arguments and read what it prints.
package cliflags

import (
	"errors"
	"flag"
	"fmt"
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
	if members < 1 {
		return nil, fmt.Errorf("can't split between %d members", members)
	}
	if total < members {
		return nil, fmt.Errorf("Ksh%d can't be split between %d members", total, members)
	}
	shares := make([]int, members)
	for i := range shares {
		shares[i] = total / members
		if i < total%members {
			shares[i]++
		}
	}
	return shares, nil
}

// Loan works out a flat-rate chama loan: interest is ratePercent of amount for every
// month, rounded down; total is amount plus interest; monthly is total divided by
// months, rounded up so the loan is fully repaid.
//
//	Loan(20000, 10, 3) → interest 6000, total 26000, monthly 8667
func Loan(amount, ratePercent, months int) (interest, total, monthly int) {
	interest = amount * ratePercent * months / 100
	total = amount + interest
	monthly = (total + months - 1) / months
	return interest, total, monthly
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
	if len(args) == 0 {
		fmt.Fprint(stderr, Usage)
		return ExitUsage
	}
	switch args[0] {
	case "split":
		return runSplit(args[1:], stdout, stderr)
	case "loan":
		return runLoan(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintf(stdout, "chama %s\n", Version)
		return ExitOK
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, Usage)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "chama: unknown command %q\n\n%s", args[0], Usage)
		return ExitUsage
	}
}

// parse parses args with fs and returns an exit code and false if Run should stop.
func parse(fs *flag.FlagSet, args []string, stderr io.Writer) (int, bool) {
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK, false
		}
		return ExitUsage, false // the flag package has already printed the problem
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "chama %s: unexpected argument %q\n", fs.Name(), fs.Arg(0))
		return ExitUsage, false
	}
	return ExitOK, true
}

// isSet reports whether the flag called name was given on the command line.
func isSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// atLeastOne checks that each named required flag was given and is at least 1.
func atLeastOne(fs *flag.FlagSet, values map[string]int, names []string, stderr io.Writer) bool {
	for _, name := range names {
		if !isSet(fs, name) {
			fmt.Fprintf(stderr, "chama %s: -%s is required\n", fs.Name(), name)
			return false
		}
		if values[name] < 1 {
			fmt.Fprintf(stderr, "chama %s: -%s must be at least 1, got %d\n", fs.Name(), name, values[name])
			return false
		}
	}
	return true
}

func runSplit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("split", flag.ContinueOnError)
	total := fs.Int("total", 0, "the amount to share, in Ksh (required)")
	members := fs.Int("members", 0, "how many members share it (required)")
	if code, ok := parse(fs, args, stderr); !ok {
		return code
	}
	values := map[string]int{"total": *total, "members": *members}
	if !atLeastOne(fs, values, []string{"total", "members"}, stderr) {
		return ExitUsage
	}

	shares, err := Split(*total, *members)
	if err != nil {
		fmt.Fprintf(stderr, "chama split: %v\n", err)
		return ExitError
	}
	for i, s := range shares {
		fmt.Fprintf(stdout, "member %d: %s\n", i+1, ksh(s))
	}
	return ExitOK
}

func runLoan(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("loan", flag.ContinueOnError)
	amount := fs.Int("amount", 0, "the loan, in Ksh (required)")
	rate := fs.Int("rate", 10, "interest in percent per month, 0 to 100")
	months := fs.Int("months", 3, "months to repay, 1 to 36")
	if code, ok := parse(fs, args, stderr); !ok {
		return code
	}
	if !atLeastOne(fs, map[string]int{"amount": *amount}, []string{"amount"}, stderr) {
		return ExitUsage
	}
	if *rate < 0 || *rate > 100 {
		fmt.Fprintf(stderr, "chama loan: -rate must be from 0 to 100, got %d\n", *rate)
		return ExitUsage
	}
	if *months < 1 || *months > 36 {
		fmt.Fprintf(stderr, "chama loan: -months must be from 1 to 36, got %d\n", *months)
		return ExitUsage
	}

	interest, total, monthly := Loan(*amount, *rate, *months)
	fmt.Fprintf(stdout, "interest: %s\ntotal: %s\nmonthly: %s\n", ksh(interest), ksh(total), ksh(monthly))
	return ExitOK
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
