package main

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"slices"
)

type count struct {
	name string
	n    int
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n, k int
	fmt.Fscan(in, &n, &k)
	sold := map[string]int{}
	for i := 0; i < n; i++ {
		var name string
		fmt.Fscan(in, &name)
		sold[name]++
	}

	list := make([]count, 0, len(sold))
	for name, c := range sold {
		list = append(list, count{name, c})
	}
	slices.SortFunc(list, func(a, b count) int {
		if a.n != b.n {
			return cmp.Compare(b.n, a.n)
		}
		return cmp.Compare(a.name, b.name)
	})
	for i := 0; i < k && i < len(list); i++ {
		fmt.Fprintln(out, list[i].name, list[i].n)
	}
}
