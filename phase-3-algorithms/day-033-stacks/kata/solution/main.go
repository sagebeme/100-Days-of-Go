package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	var s string
	fmt.Fscan(in, &n, &s)

	// The stack holds indexes of unmatched '(' on top of the index of the last
	// unmatched ')', which marks where a balanced stretch can start after.
	stack := []int{-1}
	best := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			stack = append(stack, i)
			continue
		}
		stack = stack[:len(stack)-1]
		if len(stack) == 0 {
			stack = append(stack, i) // an unmatched ')': the new floor
			continue
		}
		best = max(best, i-stack[len(stack)-1])
	}
	fmt.Println(best)
}
