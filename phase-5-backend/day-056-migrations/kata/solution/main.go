package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

// ready is a min-heap of names, smallest first.
type ready []string

func (h ready) Len() int           { return len(h) }
func (h ready) Less(i, j int) bool { return h[i] < h[j] }
func (h ready) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *ready) Push(x any)        { *h = append(*h, x.(string)) }
func (h *ready) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n, m int
	fmt.Fscan(in, &n, &m)
	names := make([]string, n)
	indegree := map[string]int{}
	next := map[string][]string{}
	for i := range names {
		fmt.Fscan(in, &names[i])
		indegree[names[i]] = 0
	}
	for i := 0; i < m; i++ {
		var a, b string
		fmt.Fscan(in, &a, &b)
		next[a] = append(next[a], b)
		indegree[b]++
	}

	h := &ready{}
	for _, name := range names {
		if indegree[name] == 0 {
			heap.Push(h, name)
		}
	}
	order := make([]string, 0, n)
	for h.Len() > 0 {
		name := heap.Pop(h).(string)
		order = append(order, name)
		for _, after := range next[name] {
			indegree[after]--
			if indegree[after] == 0 {
				heap.Push(h, after)
			}
		}
	}

	if len(order) < n {
		fmt.Fprintln(out, "CYCLE")
		return
	}
	for _, name := range order {
		fmt.Fprintln(out, name)
	}
}
