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

	// A deque of indexes whose queue lengths fall from front to back.
	// The front is always the window's maximum.
	dq := make([]int, n) // the deque lives in dq[head:tail]
	head, tail := 0, 0
	total := 0
	for i := 0; i < n; i++ {
		for tail > head && log[dq[tail-1]] <= log[i] {
			tail--
		}
		dq[tail] = i
		tail++
		if dq[head] <= i-k {
			head++
		}
		if i >= k-1 {
			total += int(log[dq[head]] - '0')
		}
	}
	fmt.Println(total)
}
