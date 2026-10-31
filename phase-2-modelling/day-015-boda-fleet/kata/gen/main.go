// Command gen prints the big case for this kata: 200 000 trips by 50 000 riders whose names are
// all 20 characters long, so the input is about as many bytes as the limits allow. Reading it one
// byte per system call (fmt.Fscan straight on os.Stdin, with no bufio.Reader) takes seconds.
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
	const n, k, riders = 200000, 1000, 50000
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	names := make([][]byte, riders)
	for i := range names {
		w := make([]byte, 20)
		for j := range w {
			w[j] = chars[r.IntN(len(chars))]
		}
		names[i] = w
	}
	out.WriteString(strconv.Itoa(n) + " " + strconv.Itoa(k) + "\n")
	for i := 0; i < n; i++ {
		out.Write(names[r.IntN(riders)])
		out.WriteString(" " + strconv.Itoa(1+r.IntN(100)) + " " + strconv.Itoa(r.IntN(10001)) + "\n")
	}
}
