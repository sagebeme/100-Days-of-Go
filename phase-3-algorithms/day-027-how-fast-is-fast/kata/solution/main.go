package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	var log string
	fmt.Fscan(in, &n, &log)

	// One pass: every W pairs with all the Ds seen before it. O(n).
	deposits, pairs := 0, 0
	for i := 0; i < len(log); i++ {
		switch log[i] {
		case 'D':
			deposits++
		case 'W':
			pairs += deposits
		}
	}
	fmt.Println(pairs)
}
