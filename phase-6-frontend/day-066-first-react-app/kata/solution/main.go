package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	children := map[string][]string{}
	root := ""
	for i := 0; i < n; i++ {
		var name, parent string
		fmt.Fscan(in, &name, &parent)
		if parent == "-" {
			root = name
			continue
		}
		children[parent] = append(children[parent], name)
	}

	maxDepth, leaves := 0, 0
	var walk func(name string, depth int)
	walk = func(name string, depth int) {
		fmt.Fprintf(out, "%s%s\n", strings.Repeat("  ", depth), name)
		maxDepth = max(maxDepth, depth+1)
		if len(children[name]) == 0 {
			leaves++
		}
		for _, c := range children[name] {
			walk(c, depth+1)
		}
	}
	walk(root, 0)
	fmt.Fprintf(out, "depth %d\nleaves %d\n", maxDepth, leaves)
}
