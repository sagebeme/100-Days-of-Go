// Command gen prints the big case for this kata: L = 1 000 000 lines and 100 000 blocks. Most
// blocks are about 300 000 lines long, so marking every line of every block is around 3 × 10^10
// steps. The rest are short blocks scattered through the middle of the file, which leaves
// thousands of uncovered ranges to print.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const lines, n = 1000000, 100000
	out.WriteString(strconv.Itoa(lines) + " " + strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		var l, rr int
		switch {
		case i%10 == 9: // a short block in the middle fifth
			l = 400001 + r.IntN(200000)
			rr = min(l+r.IntN(5), lines)
		case i%2 == 0: // a long block in the first part
			l = 1 + r.IntN(50000)
			rr = 350000 + r.IntN(50001)
		default: // a long block in the last part, stopping short of the end
			l = 600001 + r.IntN(50000)
			rr = 950000 + r.IntN(49990)
		}
		out.WriteString(strconv.Itoa(l) + " " + strconv.Itoa(rr) + "\n")
	}
}
