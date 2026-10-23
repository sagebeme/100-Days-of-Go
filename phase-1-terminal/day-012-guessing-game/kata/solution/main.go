package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var t int
	fmt.Fscan(in, &t)
	for i := 0; i < t; i++ {
		var n, secret int
		fmt.Fscan(in, &n, &secret)
		fmt.Fprintln(out, guesses(n, secret))
	}
}

func guesses(n, secret int) int {
	lo, hi := 1, n
	count := 0
	for {
		guess := (lo + hi) / 2
		count++
		if guess == secret {
			return count
		}
		if guess < secret {
			lo = guess + 1
		} else {
			hi = guess - 1
		}
	}
}
