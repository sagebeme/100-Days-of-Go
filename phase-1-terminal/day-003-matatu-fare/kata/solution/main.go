package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	var n int
	fmt.Fscan(in, &n)
	total := 0
	for i := 0; i < n; i++ {
		var age int
		fmt.Fscan(in, &age)
		switch {
		case age < 12:
			total += 30
		case age < 65:
			total += 60
		}
	}
	fmt.Fprintln(out, total)
	out.Flush()
}
