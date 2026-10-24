package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var n, b, r int
	fmt.Fscan(in, &n, &b, &r)
	tokens, last := b, 0
	sent, rejected, lastSent := 0, 0, 0
	for range n {
		var t int
		fmt.Fscan(in, &t)
		if steps := (t - last) / r; steps > 0 {
			tokens += steps
			last += steps * r
		}
		if tokens >= b {
			tokens = b
			last = t
		}
		if tokens == 0 {
			rejected++
			continue
		}
		tokens--
		sent++
		lastSent = t
	}
	fmt.Println(sent, rejected)
	fmt.Println(lastSent)
}
