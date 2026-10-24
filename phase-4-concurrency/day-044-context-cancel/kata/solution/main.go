package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	children := make([][]int, n+1)
	for i := 1; i <= n; i++ {
		var p int
		fmt.Fscan(in, &p)
		children[p] = append(children[p], i)
	}

	cancelled := make([]bool, n+1)
	var m int
	fmt.Fscan(in, &m)
	stack := []int{}
	for range m {
		var x int
		fmt.Fscan(in, &x)
		count := 0
		if !cancelled[x] {
			cancelled[x] = true
			stack = append(stack[:0], x)
		}
		for len(stack) > 0 {
			c := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			count++
			for _, child := range children[c] {
				if !cancelled[child] {
					cancelled[child] = true
					stack = append(stack, child)
				}
			}
		}
		fmt.Fprintln(out, count)
	}
}
