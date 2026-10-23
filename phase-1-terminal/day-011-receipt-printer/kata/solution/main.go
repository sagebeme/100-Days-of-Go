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

	var n int
	fmt.Fscan(in, &n)
	lines := make([]string, n)
	width := 0
	for i := range lines {
		var a int
		fmt.Fscan(in, &a)
		lines[i] = commas(a)
		width = max(width, len(lines[i]))
	}
	for _, s := range lines {
		fmt.Fprintf(out, "%*s\n", width, s)
	}
}

func commas(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := strconv.Itoa(n)
	b := make([]byte, 0, len(digits)+len(digits)/3)
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, digits[i])
	}
	return sign + string(b)
}
