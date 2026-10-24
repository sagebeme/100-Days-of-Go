package main

import (
	"bufio"
	"fmt"
	"os"
)

type account struct {
	fails int
	until int // locked while t < until; -1 means not locked
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	accounts := map[string]*account{}
	for i := 0; i < n; i++ {
		var t int
		var phone, result string
		fmt.Fscan(in, &t, &phone, &result)
		a, ok := accounts[phone]
		if !ok {
			a = &account{until: -1}
			accounts[phone] = a
		}
		if a.until != -1 && t < a.until {
			fmt.Fprintln(out, "LOCKED")
			continue
		}
		if a.until != -1 {
			a.fails, a.until = 0, -1
		}
		if result == "OK" {
			a.fails = 0
			fmt.Fprintln(out, "OK")
			continue
		}
		a.fails++
		if a.fails == 3 {
			a.fails, a.until = 0, t+300
			fmt.Fprintln(out, "FAIL LOCKED")
			continue
		}
		fmt.Fprintln(out, "FAIL")
	}
}
