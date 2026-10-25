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

	var n, ttl, d int
	fmt.Fscan(in, &n, &ttl, &d)
	expiry, refreshDone, refreshes := ttl, -1, 0
	for i := 0; i < n; i++ {
		var t int
		fmt.Fscan(in, &t)
		switch {
		case t < refreshDone: // a refresh is running: wait for it
			fmt.Fprintln(out, refreshDone)
		case t < expiry: // the token still works
			fmt.Fprintln(out, t)
		default: // expired: this request starts the one shared refresh
			refreshes++
			refreshDone = t + d
			expiry = refreshDone + ttl
			fmt.Fprintln(out, refreshDone)
		}
	}
	fmt.Fprintf(out, "refreshes %d\n", refreshes)
}
