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
	counts := map[string]int{}
	for i := 0; i < n; i++ {
		var w string
		fmt.Fscan(in, &w)
		counts[w]++
	}

	words := make([]string, 0, len(counts))
	for w := range counts {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		a, b := words[i], words[j]
		if counts[a] != counts[b] {
			return counts[a] > counts[b]
		}
		return a < b
	})

	for _, w := range words {
		fmt.Fprintln(out, w, counts[w])
	}
}
