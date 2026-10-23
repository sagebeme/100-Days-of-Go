package main

import (
	"bufio"
	"fmt"
	"os"
)

// update returns the new lowest and highest after seeing x.
func update(lo, hi, x int) (int, int) {
	if x < lo {
		lo = x
	}
	if x > hi {
		hi = x
	}
	return lo, hi
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	var n, first int
	fmt.Fscan(in, &n, &first)
	lo, hi := first, first
	for i := 1; i < n; i++ {
		var x int
		fmt.Fscan(in, &x)
		lo, hi = update(lo, hi, x)
	}
	fmt.Fprintln(out, lo, hi)
	out.Flush()
}
