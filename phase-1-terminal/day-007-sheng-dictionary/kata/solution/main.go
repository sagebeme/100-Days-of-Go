package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	var n int
	fmt.Fscan(in, &n)
	counts := make(map[string]int)
	for i := 0; i < n; i++ {
		var word string
		fmt.Fscan(in, &word)
		counts[word]++
	}

	best, bestCount := "", 0
	for w, c := range counts {
		if c > bestCount || (c == bestCount && w < best) {
			best, bestCount = w, c
		}
	}
	fmt.Fprintln(out, best, bestCount)
	out.Flush()
}
