package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

type worker struct{ free, id int }

// byFree is a min-heap of workers: the earliest free first, then the smallest number.
type byFree []worker

func (h byFree) Len() int { return len(h) }
func (h byFree) Less(i, j int) bool {
	if h[i].free != h[j].free {
		return h[i].free < h[j].free
	}
	return h[i].id < h[j].id
}
func (h byFree) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *byFree) Push(x any)   { *h = append(*h, x.(worker)) }
func (h *byFree) Pop() any {
	old := *h
	w := old[len(old)-1]
	*h = old[:len(old)-1]
	return w
}

func main() {
	in := bufio.NewReader(os.Stdin)

	var n, k int
	fmt.Fscan(in, &n, &k)
	h := make(byFree, k)
	for i := range h {
		h[i] = worker{free: 0, id: i + 1}
	}
	heap.Init(&h)

	jobs := make([]int, k+1)
	end := 0
	for range n {
		var t int
		fmt.Fscan(in, &t)
		h[0].free += t
		jobs[h[0].id]++
		end = max(end, h[0].free)
		heap.Fix(&h, 0)
	}

	best := 1
	for id := 2; id <= k; id++ {
		if jobs[id] > jobs[best] {
			best = id
		}
	}
	fmt.Println(end)
	fmt.Println(best, jobs[best])
}
