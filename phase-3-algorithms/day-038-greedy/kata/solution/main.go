package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n, k int
	var digits string
	fmt.Fscan(in, &n, &k, &digits)

	// Keep a stack of kept digits that never goes down. A digit bigger than
	// the one after it should be crossed out: the earlier a digit is, the
	// more it's worth. Each digit is pushed and popped at most once: O(n).
	kept := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		d := digits[i]
		for k > 0 && len(kept) > 0 && kept[len(kept)-1] > d {
			kept = kept[:len(kept)-1]
			k--
		}
		kept = append(kept, d)
	}
	kept = kept[:len(kept)-k] // still some to cross out: the end is the biggest

	result := strings.TrimLeft(string(kept), "0")
	if result == "" {
		result = "0"
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	fmt.Fprintln(out, result)
}
