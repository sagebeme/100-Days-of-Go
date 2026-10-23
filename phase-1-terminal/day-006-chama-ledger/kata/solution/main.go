package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	var n, k int
	fmt.Fscan(in, &n, &k)
	members := make([]int, n)
	for i := range members {
		fmt.Fscan(in, &members[i])
	}

	start := k % n
	for i := 0; i < n; i++ {
		if i > 0 {
			fmt.Fprint(out, " ")
		}
		fmt.Fprint(out, members[(start+i)%n])
	}
	fmt.Fprintln(out)
	out.Flush()
}
