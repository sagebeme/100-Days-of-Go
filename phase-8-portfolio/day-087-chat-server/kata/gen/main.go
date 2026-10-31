// Command gen prints the big case for this kata: the full 200 000 events, 100 000 users joining
// one room and then 100 000 messages to it from random users (a few of them never joined).
// Keeping each room as a slice and searching it for the user on every event is
// 200 000 × 100 000 steps.
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
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n/2; i++ {
		out.WriteString("JOIN user" + strconv.Itoa(i) + " route-125\n")
	}
	for i := 0; i < n/2; i++ {
		out.WriteString("SAY user" + strconv.Itoa(r.IntN(n/2+n/10)) + " route-125\n")
	}
}
