// Command gen prints the big case for this kata: the full 200 000 days with K = 0, each day 0, 1
// or -1 cents (mostly 0) and the running total kept between -2 and 2, so the same totals come
// back again and again and the answer is in the billions, past 2^31. Checking every run of days
// is 200 000 × 200 000 / 2 steps.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n = 200000
	days := []int{0, 0, 0, 1, -1}
	out.WriteString(strconv.Itoa(n) + " 0\n")
	prefix := 0
	for i := 0; i < n; i++ {
		if i > 0 {
			out.WriteByte(' ')
		}
		x := days[r.IntN(len(days))]
		if prefix+x > 2 || prefix+x < -2 {
			x = -x
		}
		prefix += x
		out.WriteString(strconv.Itoa(x))
	}
	out.WriteByte('\n')
}
