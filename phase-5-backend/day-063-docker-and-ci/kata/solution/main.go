package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strings"
)

// minHeap is a priority queue of job names, smallest first.
type minHeap []string

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(string)) }
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

	var n, m int
	fmt.Fscan(in, &n, &m)
	jobs := make([]string, n)
	waiting := map[string]int{}      // job → how many of its needs haven't run yet
	unlocks := map[string][]string{} // job → jobs that need it
	for i := range jobs {
		fmt.Fscan(in, &jobs[i])
		waiting[jobs[i]] = 0
	}
	for i := 0; i < m; i++ {
		var job, need string
		fmt.Fscan(in, &job, &need)
		waiting[job]++
		unlocks[need] = append(unlocks[need], job)
	}

	// Kahn's algorithm: run any job with nothing left to wait for, always the
	// alphabetically smallest one first.
	h := &minHeap{}
	for _, j := range jobs {
		if waiting[j] == 0 {
			heap.Push(h, j)
		}
	}
	order := make([]string, 0, n)
	for h.Len() > 0 {
		j := heap.Pop(h).(string)
		order = append(order, j)
		for _, next := range unlocks[j] {
			waiting[next]--
			if waiting[next] == 0 {
				heap.Push(h, next)
			}
		}
	}
	if len(order) < n {
		fmt.Fprintln(out, "CYCLE")
		return
	}
	fmt.Fprintln(out, strings.Join(order, " "))
}
