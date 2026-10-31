// Command gen prints the big case for this kata: the full 200 000 requests, all from one IP, one
// a second with w = k = 1000, so every request is allowed. Keeping every time ever allowed and
// counting through all of them on each request is 200 000 × 200 000 / 2 steps.
package main

import (
	"bufio"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n = 200000
	out.WriteString(strconv.Itoa(n) + " 1000 1000\n")
	for i := 0; i < n; i++ {
		out.WriteString(strconv.Itoa(i) + " 41.90.1.1\n")
	}
}
