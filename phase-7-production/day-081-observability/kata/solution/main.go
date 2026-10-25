package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

// intHeap is a min-heap, or a max-heap when max is true.
type intHeap struct {
	v   []int
	max bool
}

func (h intHeap) Len() int { return len(h.v) }
func (h intHeap) Less(i, j int) bool {
	if h.max {
		return h.v[i] > h.v[j]
	}
	return h.v[i] < h.v[j]
}
func (h intHeap) Swap(i, j int) { h.v[i], h.v[j] = h.v[j], h.v[i] }
func (h *intHeap) Push(x any)   { h.v = append(h.v, x.(int)) }
func (h *intHeap) Pop() any {
	x := h.v[len(h.v)-1]
	h.v = h.v[:len(h.v)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q, p int
	fmt.Fscan(in, &q, &p)
	low := &intHeap{max: true} // the smallest k samples; its top is the answer
	high := &intHeap{}         // everything else
	n := 0
	for range q {
		var op string
		fmt.Fscan(in, &op)
		if op == "query" {
			if n == 0 {
				fmt.Fprintln(out, "-")
			} else {
				fmt.Fprintln(out, low.v[0])
			}
			continue
		}
		var x int
		fmt.Fscan(in, &x)
		n++
		if low.Len() > 0 && x <= low.v[0] {
			heap.Push(low, x)
		} else {
			heap.Push(high, x)
		}
		k := (p*n + 99) / 100
		for low.Len() < k {
			heap.Push(low, heap.Pop(high))
		}
		for low.Len() > k {
			heap.Push(high, heap.Pop(low))
		}
	}
}
