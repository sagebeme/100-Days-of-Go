package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

// version is one change to a key. deleted marks a DEL.
type version struct {
	t       int
	value   string
	deleted bool
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q int
	fmt.Fscan(in, &q)
	history := map[string][]version{}
	for range q {
		var cmd, key string
		fmt.Fscan(in, &cmd, &key)
		switch cmd {
		case "SET":
			var v version
			fmt.Fscan(in, &v.value, &v.t)
			history[key] = append(history[key], v) // times only grow, so it stays sorted
		case "DEL":
			var t int
			fmt.Fscan(in, &t)
			history[key] = append(history[key], version{t: t, deleted: true})
		case "GET":
			var t int
			fmt.Fscan(in, &t)
			h := history[key]
			// The first version after t; the one before it is the answer.
			i := sort.Search(len(h), func(i int) bool { return h[i].t > t }) - 1
			if i < 0 || h[i].deleted {
				fmt.Fprintln(out, "-")
			} else {
				fmt.Fprintln(out, h[i].value)
			}
		}
	}
}
