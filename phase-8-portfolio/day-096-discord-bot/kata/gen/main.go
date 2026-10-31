// Command gen prints the big case for this kata: 150 000 lines, 1 000 TOP lines evenly spread and
// the rest solves by 1 000 users whose names are the full 20 characters, about 4.5 MB of input.
// Reading it a byte at a time means millions of system calls. (The full 200 000 lines would push
// the reference close to the limit on a slow machine.)
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
	const q, tops = 150000, 1000
	users := make([]string, 1000)
	for i := range users {
		// "kenyadevuser" and an 8-digit number: 20 characters.
		users[i] = "kenyadevuser" + strconv.Itoa(10000000+r.IntN(90000000))
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		if i%(q/tops) == 0 {
			out.WriteString("TOP " + strconv.Itoa(1+r.IntN(10)) + "\n")
			continue
		}
		out.WriteString("SOLVE " + users[r.IntN(len(users))] + " " + strconv.Itoa(1+r.IntN(100)) + " " + strconv.Itoa(r.IntN(11)) + "\n")
	}
}
