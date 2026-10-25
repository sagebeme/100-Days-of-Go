package main

import (
	"bufio"
	"fmt"
	"os"
)

type key struct{ month, category string }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var b int
	fmt.Fscan(in, &b)
	budgets := map[string]int{}
	for i := 0; i < b; i++ {
		var cat string
		var amount int
		fmt.Fscan(in, &cat, &amount)
		budgets[cat] = amount
	}

	var n int
	fmt.Fscan(in, &n)
	spent := map[key]int{}
	alerted := map[key]bool{}
	alerts := 0
	for i := 0; i < n; i++ {
		var date, cat string
		var amount int
		fmt.Fscan(in, &date, &cat, &amount)
		k := key{date[:7], cat}
		spent[k] += amount
		budget, ok := budgets[cat]
		if ok && spent[k] > budget && !alerted[k] {
			alerted[k] = true
			alerts++
			fmt.Fprintln(out, date, cat, spent[k], budget)
		}
	}
	if alerts == 0 {
		fmt.Fprintln(out, "no alerts")
	}
}
