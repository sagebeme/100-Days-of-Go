package main

import (
	"bufio"
	"fmt"
	"os"
)

type seat struct {
	trip string
	n    int
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	held := map[seat]string{}   // seat → the booking holding it
	active := map[string]seat{} // booking → its seat, while it's active
	for i := 0; i < n; i++ {
		var op, id string
		fmt.Fscan(in, &op, &id)
		if op == "BOOK" {
			var s seat
			fmt.Fscan(in, &s.trip, &s.n)
			if holder, taken := held[s]; taken {
				fmt.Fprintln(out, "TAKEN", holder)
				continue
			}
			held[s] = id
			active[id] = s
			fmt.Fprintln(out, "OK")
			continue
		}
		s, ok := active[id]
		if !ok {
			fmt.Fprintln(out, "UNKNOWN")
			continue
		}
		delete(active, id)
		delete(held, s)
		fmt.Fprintln(out, "CANCELLED")
	}
	fmt.Fprintln(out, len(active))
}
