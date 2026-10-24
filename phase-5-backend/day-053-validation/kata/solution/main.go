package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func valid(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) < 1 || len(p) > 3 {
			return false
		}
		for i := 0; i < len(p); i++ {
			if p[i] < '0' || p[i] > '9' {
				return false
			}
		}
		if len(p) > 1 && p[0] == '0' {
			return false
		}
		if n, _ := strconv.Atoi(p); n > 255 {
			return false
		}
	}
	return true
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	for i := 0; i < n; i++ {
		line := ""
		if sc.Scan() {
			line = sc.Text()
		}
		if valid(line) {
			fmt.Fprintln(out, "VALID")
		} else {
			fmt.Fprintln(out, "INVALID")
		}
	}
}
