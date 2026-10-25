package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strings"
)

type route struct {
	name  string
	total int
}

// worse reports whether a ranks below b: fewer seats, or the same and a later name.
func worse(a, b route) bool {
	if a.total != b.total {
		return a.total < b.total
	}
	return a.name > b.name
}

// minHeap keeps the worst of the routes it holds at the top.
type minHeap []route

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return worse(h[i], h[j]) }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(route)) }
func (h *minHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	totals := map[string]int{}
	for range n {
		var op string
		fmt.Fscan(in, &op)
		if op == "book" {
			var name string
			var seats int
			fmt.Fscan(in, &name, &seats)
			totals[name] += seats
			continue
		}
		var k int
		fmt.Fscan(in, &k)
		h := &minHeap{}
		for name, total := range totals {
			heap.Push(h, route{name, total})
			if h.Len() > k {
				heap.Pop(h)
			}
		}
		if h.Len() == 0 {
			fmt.Fprintln(out, "-")
			continue
		}
		best := make([]string, h.Len())
		for i := len(best) - 1; i >= 0; i-- {
			best[i] = heap.Pop(h).(route).name
		}
		fmt.Fprintln(out, strings.Join(best, " "))
	}
}
