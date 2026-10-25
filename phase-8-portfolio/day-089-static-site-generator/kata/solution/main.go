package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func check(line string) string {
	var stack []string
	for i := 0; i < len(line); i++ {
		if line[i] != '<' {
			continue
		}
		end := strings.IndexByte(line[i:], '>') + i
		tag := line[i+1 : end]
		switch {
		case strings.HasSuffix(tag, "/"):
			// self-closing: nothing to match
		case strings.HasPrefix(tag, "/"):
			if len(stack) == 0 || stack[len(stack)-1] != tag[1:] {
				return "MISMATCH " + strconv.Itoa(i+1)
			}
			stack = stack[:len(stack)-1]
		default:
			stack = append(stack, tag)
		}
		i = end
	}
	if len(stack) > 0 {
		return "UNCLOSED " + stack[len(stack)-1]
	}
	return "OK"
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 1<<20), 1<<20)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	in.Scan()
	q, _ := strconv.Atoi(strings.TrimSpace(in.Text()))
	for range q {
		in.Scan() // at the end of input, Text is "" which is a valid empty line
		fmt.Fprintln(out, check(in.Text()))
	}
}
