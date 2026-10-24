package main

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
)

type entry struct{ key, value int }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var capacity, q int
	fmt.Fscan(in, &capacity, &q)
	order := list.New() // front is the most recently used
	items := make(map[int]*list.Element, capacity)
	for range q {
		var op string
		var key int
		fmt.Fscan(in, &op, &key)
		if op == "get" {
			e, ok := items[key]
			if !ok {
				fmt.Fprintln(out, -1)
				continue
			}
			order.MoveToFront(e)
			fmt.Fprintln(out, e.Value.(*entry).value)
			continue
		}
		var value int
		fmt.Fscan(in, &value)
		if e, ok := items[key]; ok {
			e.Value.(*entry).value = value
			order.MoveToFront(e)
			continue
		}
		if order.Len() == capacity {
			oldest := order.Back()
			order.Remove(oldest)
			delete(items, oldest.Value.(*entry).key)
		}
		items[key] = order.PushFront(&entry{key, value})
	}
}
