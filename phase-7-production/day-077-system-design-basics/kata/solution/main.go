package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"sort"
)

type point struct {
	pos  int
	name string
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q int
	fmt.Fscan(in, &q)
	var ring []point // sorted by pos
	for range q {
		var op string
		fmt.Fscan(in, &op)
		switch op {
		case "add":
			var p point
			fmt.Fscan(in, &p.name, &p.pos)
			i := sort.Search(len(ring), func(i int) bool { return ring[i].pos >= p.pos })
			ring = slices.Insert(ring, i, p)
		case "remove":
			var name string
			fmt.Fscan(in, &name)
			ring = slices.DeleteFunc(ring, func(p point) bool { return p.name == name })
		case "get":
			var h int
			fmt.Fscan(in, &h)
			if len(ring) == 0 {
				fmt.Fprintln(out, "NONE")
				continue
			}
			i := sort.Search(len(ring), func(i int) bool { return ring[i].pos >= h })
			if i == len(ring) {
				i = 0
			}
			fmt.Fprintln(out, ring[i].name)
		}
	}
}
