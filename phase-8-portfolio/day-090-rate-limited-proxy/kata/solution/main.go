package main

import (
	"bufio"
	"fmt"
	"os"
)

type bucket struct{ milli, last int }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var b, r, n int
	fmt.Fscan(in, &b, &r, &n)
	capacity := b * 1000
	buckets := map[string]bucket{}
	allowed := 0
	for range n {
		var t int
		var client string
		fmt.Fscan(in, &t, &client)
		bk, ok := buckets[client]
		if !ok {
			bk = bucket{milli: capacity, last: t}
		}
		bk.milli = min(capacity, bk.milli+r*(t-bk.last))
		bk.last = t
		if bk.milli >= 1000 {
			bk.milli -= 1000
			allowed++
			fmt.Fprintln(out, "ALLOW")
		} else {
			fmt.Fprintln(out, "DENY")
		}
		buckets[client] = bk
	}
	fmt.Fprintln(out, allowed)
}
