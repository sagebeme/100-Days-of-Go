// Command gen prints the big case for this kata: the full 200 000 words, every one 20 letters
// long, so the input is as many bytes as the limits allow. The words come from a pool of 40 000,
// so many are tied near the top and the alphabetical tie rule decides the answer. Reading it one
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
	const n, pool = 200000, 40000
	words := make([][]byte, pool)
	for i := range words {
		w := make([]byte, 20)
		for j := range w {
			w[j] = byte('a' + r.IntN(26))
		}
		words[i] = w
	}
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		if i > 0 {
			if i%10 == 0 {
				out.WriteByte('\n')
			} else {
				out.WriteByte(' ')
			}
		}
		out.Write(words[r.IntN(pool)])
	}
	out.WriteByte('\n')
}
