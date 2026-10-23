package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	for i := 0; i < n; i++ {
		sc.Scan()
		if valid(sc.Text()) {
			fmt.Fprintln(out, "VALID")
		} else {
			fmt.Fprintln(out, "INVALID")
		}
	}
}

func valid(code string) bool {
	if len(code) != 10 {
		return false
	}
	if code[0] < 'A' || code[0] > 'Z' {
		return false
	}
	hasDigit := false
	for i := 0; i < len(code); i++ {
		c := code[i]
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			return false
		}
	}
	return hasDigit
}
