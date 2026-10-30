package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	runs := map[string]int{}
	passes := map[string]int{}
	for i := 0; i < n; i++ {
		var name, result string
		fmt.Fscan(in, &name, &result)
		runs[name]++
		if result == "PASS" {
			passes[name]++
		}
	}

	names := make([]string, 0, len(runs))
	for name := range runs {
		names = append(names, name)
	}
	sort.Strings(names)

	stable := 0
	for _, name := range names {
		switch p, t := passes[name], runs[name]; {
		case p == t:
			stable++
		case p == 0:
			fmt.Fprintf(out, "broken %s\n", name)
		default:
			fmt.Fprintf(out, "flaky %s %d/%d\n", name, p, t)
		}
	}
	fmt.Fprintf(out, "stable %d\n", stable)
}
