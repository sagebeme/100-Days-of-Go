// Command gen prints the big case for this kata: the full 200 000 customers with 9-digit arrival
// minutes and patience that is either 9 or 10 digits or a few thousand minutes, so the input is
// close to the most bytes the limits allow. The teller is overloaded (a customer arrives every
// 500 minutes on average and needs 5000), so a long queue builds and many give up. Reading it one
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
	const n = 200000
	out.WriteString(strconv.Itoa(n) + "\n")
	a := 900000000
	for i := 0; i < n; i++ {
		a += r.IntN(1000) // ends near 1 000 000 000, never past it
		s := 1 + r.IntN(10000)
		var p int
		switch x := r.IntN(10); {
		case x < 4:
			p = 100000000 + r.IntN(900000001)
		case x < 5:
			p = 1000000000
		default:
			p = r.IntN(20001)
		}
		out.WriteString(strconv.Itoa(a) + " " + strconv.Itoa(s) + " " + strconv.Itoa(p) + "\n")
	}
}
