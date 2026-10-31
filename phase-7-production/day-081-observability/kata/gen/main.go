// Command gen prints the big case for this kata: the full 200 000 operations with p = 95, seven
// adds to every query, and each latency smaller than the one before it. Sorting all the samples
// again on every query is 25 000 sorts of up to 175 000 numbers, and keeping one sorted slice
// with an insert is no better here: every new sample goes in at the front, so each insert shifts
// the whole slice.
package main

import (
	"bufio"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const q = 200000
	out.WriteString(strconv.Itoa(q) + " 95\n")
	ms := 1000000000
	for i := 0; i < q; i++ {
		if i%8 == 7 {
			out.WriteString("query\n")
			continue
		}
		ms -= 1 + i%7*500
		out.WriteString("add " + strconv.Itoa(ms) + "\n")
	}
}
