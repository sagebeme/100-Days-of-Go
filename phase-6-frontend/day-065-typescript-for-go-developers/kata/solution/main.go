package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	totals := map[string]int{}
	refunds, failed := 0, 0
	for i := 0; i < n; i++ {
		var kind string
		fmt.Fscan(in, &kind)
		switch kind {
		case "mpesa", "card":
			var amount int
			var extra string // the phone number or the card's last four digits
			fmt.Fscan(in, &amount, &extra)
			totals[kind] += amount
		case "cash":
			var amount int
			fmt.Fscan(in, &amount)
			totals[kind] += amount
		case "refund":
			var amount int
			fmt.Fscan(in, &amount)
			refunds += amount
		case "failed":
			var reason string
			fmt.Fscan(in, &reason)
			failed++
		}
	}
	net := totals["mpesa"] + totals["cash"] + totals["card"] - refunds
	fmt.Fprintf(out, "mpesa %d\ncash %d\ncard %d\nrefunds %d\nnet %d\nfailed %d\n",
		totals["mpesa"], totals["cash"], totals["card"], refunds, net, failed)
}
