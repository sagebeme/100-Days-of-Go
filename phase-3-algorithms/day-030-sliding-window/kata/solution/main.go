package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n, k int
	var log string
	fmt.Fscan(in, &n, &k, &log)

	// Grow the window on the right. When it holds more than k skipped days,
	// shrink it from the left until it's allowed again. Each day enters and
	// leaves the window once: O(n).
	best, skipped, left := 0, 0, 0
	for right := 0; right < len(log); right++ {
		if log[right] == '.' {
			skipped++
		}
		for skipped > k {
			if log[left] == '.' {
				skipped--
			}
			left++
		}
		best = max(best, right-left+1)
	}
	fmt.Println(best)
}
