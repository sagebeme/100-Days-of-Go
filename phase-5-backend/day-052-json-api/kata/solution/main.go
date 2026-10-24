package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// check returns 0 if line is balanced, or the 1-based position of the first problem.
func check(line string) int {
	stack := []byte{}
	inString, escaped := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			stack = append(stack, c)
		case '}', ']':
			open := byte('{')
			if c == ']' {
				open = '['
			}
			if len(stack) == 0 || stack[len(stack)-1] != open {
				return i + 1
			}
			stack = stack[:len(stack)-1]
		}
	}
	if inString || len(stack) > 0 {
		return len(line) + 1
	}
	return 0
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
		if k := check(line); k == 0 {
			fmt.Fprintln(out, "YES")
		} else {
			fmt.Fprintln(out, "NO", k)
		}
	}
}
