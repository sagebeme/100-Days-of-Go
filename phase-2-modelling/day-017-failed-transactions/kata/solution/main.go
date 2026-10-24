package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

var (
	errInvalid = errors.New("INVALID_AMOUNT")
	errUnknown = errors.New("UNKNOWN_ACCOUNT")
	errSame    = errors.New("SAME_ACCOUNT")
	errFunds   = errors.New("INSUFFICIENT_FUNDS")
)

func check(balances map[string]int, from, to string, amount int) error {
	if amount < 1 {
		return errInvalid
	}
	_, okFrom := balances[from]
	_, okTo := balances[to]
	if !okFrom || !okTo {
		return errUnknown
	}
	if from == to {
		return errSame
	}
	if balances[from] < amount {
		return errFunds
	}
	return nil
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	names := make([]string, n)
	balances := make(map[string]int, n)
	for i := range names {
		var bal int
		fmt.Fscan(in, &names[i], &bal)
		balances[names[i]] = bal
	}

	var q int
	fmt.Fscan(in, &q)
	for ; q > 0; q-- {
		var code, from, to string
		var amount int
		fmt.Fscan(in, &code, &from, &to, &amount)
		if err := check(balances, from, to, amount); err != nil {
			fmt.Fprintln(out, code, "FAILED", err.Error())
			continue
		}
		balances[from] -= amount
		balances[to] += amount
		fmt.Fprintln(out, code, "OK")
	}
	for _, name := range names {
		fmt.Fprintln(out, name, balances[name])
	}
}
