package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// reach returns every module reachable from start along edges.
func reach(start string, edges map[string][]string) map[string]bool {
	seen := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range edges[cur] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return seen
}

func main() {
	in := bufio.NewReader(os.Stdin)
	var n, m int
	fmt.Fscan(in, &n, &m)
	size := make(map[string]int, n)
	for i := 0; i < n; i++ {
		var name string
		var bytes int
		fmt.Fscan(in, &name, &bytes)
		size[name] = bytes
	}
	static := map[string][]string{}
	all := map[string][]string{}
	for i := 0; i < m; i++ {
		var from, to, kind string
		fmt.Fscan(in, &from, &to, &kind)
		all[from] = append(all[from], to)
		if kind == "static" {
			static[from] = append(static[from], to)
		}
	}

	initial := reach("main", static)
	reachable := reach("main", all)
	initialBytes, laterBytes := 0, 0
	var unused []string
	for name, bytes := range size {
		switch {
		case initial[name]:
			initialBytes += bytes
		case reachable[name]:
			laterBytes += bytes
		default:
			unused = append(unused, name)
		}
	}
	sort.Strings(unused)
	list := "-"
	if len(unused) > 0 {
		list = strings.Join(unused, " ")
	}
	fmt.Printf("initial %d\nmodules %d\nlater %d\nunused %s\n", initialBytes, len(initial), laterBytes, list)
}
