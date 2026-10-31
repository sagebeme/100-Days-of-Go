// Command gen prints the big case for this kata: the full 200 000 sales of items whose names are
// all 20 characters long, so the input is as many bytes as the limits allow. The names come from
// a pool of 50 000, so the top 1000 ends in a long run of ties that the alphabetical rule
// settles. Reading it one byte per system call (fmt.Fscan straight on os.Stdin, with no
// bufio.Reader) takes seconds.
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
	const n, k, pool = 200000, 1000, 50000
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	names := make([][]byte, pool)
	for i := range names {
		w := make([]byte, 20)
		for j := range w {
			w[j] = chars[r.IntN(len(chars))]
		}
		names[i] = w
	}
	out.WriteString(strconv.Itoa(n) + " " + strconv.Itoa(k) + "\n")
	for i := 0; i < n; i++ {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.Write(names[r.IntN(pool)])
	}
	out.WriteByte('\n')
}
