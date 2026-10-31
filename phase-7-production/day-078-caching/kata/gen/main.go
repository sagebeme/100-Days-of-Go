// Command gen prints the big case for this kata: capacity 100 000 and 200 000 operations. The
// first 100 000 fill the cache, then 80% are puts of new keys (each one evicts) and 20% are gets.
// Scanning the whole cache for the oldest key on every eviction is 80 000 × 100 000 steps.
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
	const c, q = 100000, 200000
	out.WriteString(strconv.Itoa(c) + " " + strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		switch {
		case i < c:
			out.WriteString("put " + strconv.Itoa(i+1) + " " + strconv.Itoa(i) + "\n")
		case r.Float64() < 0.8:
			out.WriteString("put " + strconv.Itoa(1+r.IntN(1000000000)) + " " + strconv.Itoa(i) + "\n")
		default:
			out.WriteString("get " + strconv.Itoa(1+r.IntN(2*c)) + "\n")
		}
	}
}
