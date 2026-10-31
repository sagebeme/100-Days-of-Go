// Command gen prints the big case for this kata: the full limits, with every context a child of
// the one before it, so walking the subtree again on every cancel is 200 000 × 100 000 steps.
package main

import (
	"bufio"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n = 200000
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(strconv.Itoa(i))
	}
	out.WriteString("\n" + strconv.Itoa(n) + "\n")
	// Cancel the root of the chain, then cancel halfway down it over and over: a solution that
	// stops at contexts that are already cancelled answers each of those at once.
	for i := 0; i < n; i++ {
		x := n / 2
		if i%2 == 0 {
			x = 1
		}
		out.WriteString(strconv.Itoa(x) + "\n")
	}
}
