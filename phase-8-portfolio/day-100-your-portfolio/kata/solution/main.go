package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n, h int
	fmt.Fscan(in, &n, &h)

	// best[c] is the most impact from the tasks seen so far using at most c hours.
	// Going through c from high to low means each task is used at most once.
	best := make([]int64, h+1)
	for range n {
		var hours int
		var impact int64
		fmt.Fscan(in, &hours, &impact)
		for c := h; c >= hours; c-- {
			best[c] = max(best[c], best[c-hours]+impact)
		}
	}
	// best only grows with c, so the first c that reaches best[h] is the fewest hours.
	fewest := 0
	for best[fewest] != best[h] {
		fewest++
	}
	fmt.Println(best[h])
	fmt.Println(fewest)
}
