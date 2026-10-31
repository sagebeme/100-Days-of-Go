// Command gen prints the big case for this kata: the full 200 000 rules and 200 000 questions over
// 100 roles, where 150 000 of the rules are INHERIT lines, so every role reaches every other one
// through a dense web of parents. A breadth-first search per question walks up to 150 000 edges
// each time (most of all for the actions nobody was granted), 200 000 times over.
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
	const n, q, users = 200000, 200000, 10000
	role := func() string { return "role" + strconv.Itoa(r.IntN(100)) }
	user := func() string { return "u" + strconv.Itoa(r.IntN(users)) }
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		x := r.Float64()
		switch {
		case x < 0.75:
			out.WriteString("INHERIT " + role() + " " + role() + "\n")
		case x < 0.95:
			out.WriteString("GRANT " + role() + " act" + strconv.Itoa(r.IntN(5000)) + "\n")
		default:
			out.WriteString("USER " + user() + " " + role() + "\n")
		}
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		out.WriteString(user() + " act" + strconv.Itoa(r.IntN(6000)) + "\n")
	}
}
