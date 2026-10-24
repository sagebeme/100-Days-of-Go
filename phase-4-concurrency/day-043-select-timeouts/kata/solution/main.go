package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var n int
	fmt.Fscan(in, &n)
	free, served, left := 0, 0, 0
	for range n {
		var a, s, p int
		fmt.Fscan(in, &a, &s, &p)
		start := max(a, free)
		if start-a > p {
			left++
			continue
		}
		served++
		free = start + s
	}
	fmt.Println(served, left)
	fmt.Println(free)
}
