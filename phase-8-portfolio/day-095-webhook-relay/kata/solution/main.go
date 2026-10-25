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

	var n, w int64
	fmt.Fscan(in, &n, &w)
	accepted := map[string]int64{} // id -> time it was last accepted
	for range n {
		var t int64
		var id string
		fmt.Fscan(in, &t, &id)
		if last, ok := accepted[id]; ok && t-last < w {
			fmt.Fprintln(out, "DUPLICATE") // a duplicate doesn't restart the window
			continue
		}
		accepted[id] = t
		fmt.Fprintln(out, "ACCEPT")
	}
}
