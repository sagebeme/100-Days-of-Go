package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	d := make([]int, n)
	for i := range d {
		fmt.Fscan(in, &d[i])
	}
	sort.Ints(d)

	var q int
	fmt.Fscan(in, &q)
	for ; q > 0; q-- {
		var p int
		fmt.Fscan(in, &p)
		rank := (p*n + 99) / 100
		fmt.Fprintf(out, "p%d %d\n", p, d[rank-1])
	}
}
