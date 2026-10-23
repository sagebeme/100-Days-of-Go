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
	total := 0
	for i := 0; i < n; i++ {
		var mb int
		fmt.Fscan(in, &mb)
		total += mb
	}
	fmt.Fprintln(out, total)
	fmt.Fprintf(out, "%.2f\n", float64(total)/float64(n))
	out.Flush()
}
