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

	var n, k int
	fmt.Fscan(in, &n, &k)
	sequential, batched, slowest := 0, 0, 0
	for i := 0; i < n; i++ {
		var t int
		fmt.Fscan(in, &t)
		sequential += t
		slowest = max(slowest, t)
		if i%k == k-1 || i == n-1 {
			batched += slowest
			slowest = 0
		}
	}
	fmt.Fprintln(out, sequential)
	fmt.Fprintln(out, batched)
}
