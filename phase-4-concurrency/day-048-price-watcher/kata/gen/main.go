// Command gen prints the big case for this kata: the full 200 000 operations, with 100 000 shops
// all selling one item and then 90 000 `best` questions about it (a random shop stops selling it
// every tenth line), so scanning every shop on every `best` is 90 000 × 100 000 steps.
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
	const shops = 100000
	out.WriteString(strconv.Itoa(2*shops) + "\n")
	for i := 0; i < shops; i++ {
		out.WriteString("set s" + strconv.Itoa(i) + " unga " + strconv.Itoa(1+r.IntN(1000000000)) + "\n")
	}
	for i := 0; i < shops; i++ {
		if i%10 == 0 {
			out.WriteString("del s" + strconv.Itoa(r.IntN(shops)) + " unga\n")
		} else {
			out.WriteString("best unga\n")
		}
	}
}
