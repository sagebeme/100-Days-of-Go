package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

type offer struct {
	price int
	shop  string
}

// offers is a min-heap: the lowest price first, then the shop name.
type offers []offer

func (h offers) Len() int { return len(h) }
func (h offers) Less(i, j int) bool {
	if h[i].price != h[j].price {
		return h[i].price < h[j].price
	}
	return h[i].shop < h[j].shop
}
func (h offers) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *offers) Push(x any)   { *h = append(*h, x.(offer)) }
func (h *offers) Pop() any {
	old := *h
	o := old[len(old)-1]
	*h = old[:len(old)-1]
	return o
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	current := map[[2]string]int{} // {shop, item} → price
	boards := map[string]*offers{} // item → every offer ever set, some stale

	var q int
	fmt.Fscan(in, &q)
	for range q {
		var op, a string
		fmt.Fscan(in, &op, &a)
		switch op {
		case "set":
			var item string
			var price int
			fmt.Fscan(in, &item, &price)
			current[[2]string{a, item}] = price
			h := boards[item]
			if h == nil {
				h = &offers{}
				boards[item] = h
			}
			heap.Push(h, offer{price, a})
		case "del":
			var item string
			fmt.Fscan(in, &item)
			delete(current, [2]string{a, item})
		case "best":
			h := boards[a]
			for h != nil && h.Len() > 0 {
				top := (*h)[0]
				if p, ok := current[[2]string{top.shop, a}]; ok && p == top.price {
					break
				}
				heap.Pop(h) // stale: the price changed or the shop stopped selling
			}
			if h == nil || h.Len() == 0 {
				fmt.Fprintln(out, "none")
				continue
			}
			fmt.Fprintln(out, (*h)[0].shop, (*h)[0].price)
		}
	}
}
