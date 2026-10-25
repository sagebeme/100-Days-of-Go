package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

type item struct{ ready, idx int }

// readyHeap is a min-heap by ready time, then by input order.
type readyHeap []item

func (h readyHeap) Len() int { return len(h) }
func (h readyHeap) Less(i, j int) bool {
	if h[i].ready != h[j].ready {
		return h[i].ready < h[j].ready
	}
	return h[i].idx < h[j].idx
}
func (h readyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *readyHeap) Push(x any)   { *h = append(*h, x.(item)) }
func (h *readyHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n, m, base, maxWait int
	fmt.Fscan(in, &n, &m, &base, &maxWait)
	ids := make([]string, n)
	failsLeft := make([]int, n)
	failed := make([]int, n)
	result := make([]string, n)
	h := make(readyHeap, n)
	for i := range n {
		var arrive int
		fmt.Fscan(in, &ids[i], &arrive, &failsLeft[i])
		h[i] = item{arrive, i}
	}
	heap.Init(&h)

	clock := 0
	for h.Len() > 0 {
		it := heap.Pop(&h).(item)
		i := it.idx
		clock = max(clock, it.ready) + 1
		if failsLeft[i] == 0 {
			result[i] = fmt.Sprintf("%s DONE %d", ids[i], clock)
			continue
		}
		failsLeft[i]--
		failed[i]++
		if failed[i] >= m {
			result[i] = fmt.Sprintf("%s DEAD %d", ids[i], clock)
			continue
		}
		wait := base
		for k := 1; k < failed[i] && wait < maxWait; k++ {
			wait *= 2
		}
		heap.Push(&h, item{clock + min(wait, maxWait), i})
	}
	for _, r := range result {
		fmt.Fprintln(out, r)
	}
}
