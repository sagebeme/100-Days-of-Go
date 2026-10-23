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
	names := make([]string, n)
	scores := make([]int, n)
	order := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Fscan(in, &names[i], &scores[i])
		order[i] = i
	}

	sort.Slice(order, func(a, b int) bool {
		x, y := order[a], order[b]
		if scores[x] != scores[y] {
			return scores[x] > scores[y]
		}
		return names[x] < names[y]
	})

	for _, i := range order {
		fmt.Fprintln(out, names[i], scores[i])
	}
}
