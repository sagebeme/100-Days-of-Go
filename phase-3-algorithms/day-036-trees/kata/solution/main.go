package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var tree string
	fmt.Fscan(in, &tree)

	pos, best := 0, 0
	// height reads one subtree starting at tree[pos] and returns its height in
	// nodes. On the way it records the longest path that bends at each person:
	// the left height plus the right height, counted in edges.
	var height func() int
	height = func() int {
		c := tree[pos]
		pos++
		if c == '.' {
			return 0
		}
		left := height()
		right := height()
		best = max(best, left+right)
		return 1 + max(left, right)
	}
	height()
	fmt.Println(best)
}
