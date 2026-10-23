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
	current, best := 0, 0
	for i := 0; i < n; i++ {
		var mm int
		fmt.Fscan(in, &mm)
		if mm == 0 {
			current++
			if current > best {
				best = current
			}
		} else {
			current = 0
		}
	}
	fmt.Fprintln(out, best)
	out.Flush()
}
