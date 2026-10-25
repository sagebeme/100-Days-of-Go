package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

// freeAt is a min-heap of the times each connection becomes free.
type freeAt []int

func (h freeAt) Len() int           { return len(h) }
func (h freeAt) Less(i, j int) bool { return h[i] < h[j] }
func (h freeAt) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *freeAt) Push(x any)        { *h = append(*h, x.(int)) }
func (h *freeAt) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var k, n int
	fmt.Fscan(in, &k, &n)
	h := make(freeAt, k) // every connection is free at time 0
	heap.Init(&h)

	maxWait := 0
	for i := 0; i < n; i++ {
		var t, d int
		fmt.Fscan(in, &t, &d)
		start := max(t, heap.Pop(&h).(int))
		maxWait = max(maxWait, start-t)
		heap.Push(&h, start+d)
		fmt.Fprintln(out, start+d)
	}
	fmt.Fprintf(out, "max wait %d\n", maxWait)
}
