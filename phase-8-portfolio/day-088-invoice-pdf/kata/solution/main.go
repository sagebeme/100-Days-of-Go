package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var h, n int
	fmt.Fscan(in, &h, &n)
	fmt.Fprintln(out, "xref")
	fmt.Fprintln(out, "0", n+1)
	fmt.Fprintln(out, "0000000000 65535 f")
	pos := h
	for i := 1; i <= n; i++ {
		var l int
		fmt.Fscan(in, &l)
		fmt.Fprintf(out, "%010d 00000 n\n", pos)
		pos += len(strconv.Itoa(i)) + len(" 0 obj\n") + l + len("\nendobj\n")
	}
	fmt.Fprintln(out, "startxref")
	fmt.Fprintln(out, pos)
}
