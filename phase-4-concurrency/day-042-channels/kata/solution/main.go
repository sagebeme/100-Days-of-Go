package main

import (
	"bufio"
	"fmt"
	"os"
)

type customer struct {
	name string
	left int
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n, q int
	fmt.Fscan(in, &n, &q)
	queue := make([]customer, n)
	for i := range queue {
		fmt.Fscan(in, &queue[i].name, &queue[i].left)
	}

	now := 0
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c.left <= q {
			now += c.left
			fmt.Fprintln(out, c.name, now)
			continue
		}
		now += q
		c.left -= q
		queue = append(queue, c)
	}
}
