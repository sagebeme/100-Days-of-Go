// Command gen prints the big case for this kata: the full 200 000 commands, 100 000 changes to one
// key (every tenth a DEL) and then 100 000 GETs at random times. Walking the key's whole history on
// every GET is 100 000 × 100 000 steps.
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
	const q = 200000
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q/2; i++ {
		if i%10 == 9 {
			out.WriteString("DEL float " + strconv.Itoa(i+1) + "\n")
			continue
		}
		out.WriteString("SET float v" + strconv.Itoa(i) + " " + strconv.Itoa(i+1) + "\n")
	}
	for i := 0; i < q/2; i++ {
		out.WriteString("GET float " + strconv.Itoa(1+r.IntN(q)) + "\n")
	}
}
