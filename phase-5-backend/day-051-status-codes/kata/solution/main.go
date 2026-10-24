package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	var classes [6]int
	broken := map[string]int{}
	for i := 0; i < n; i++ {
		var method, path string
		var status, ms int
		fmt.Fscan(in, &method, &path, &status, &ms)
		classes[status/100]++
		if status/100 == 5 {
			broken[path]++
		}
	}

	for c := 2; c <= 5; c++ {
		fmt.Fprintf(out, "%dxx %d\n", c, classes[c])
	}
	v := classes[5] * 10000 / n
	fmt.Fprintf(out, "5xx rate %d.%02d%%\n", v/100, v%100)

	worst, most := "", 0
	for path, k := range broken {
		if k > most || (k == most && path < worst) {
			worst, most = path, k
		}
	}
	if most == 0 {
		fmt.Fprintln(out, "worst none")
	} else {
		fmt.Fprintf(out, "worst %s %d\n", worst, most)
	}
}
