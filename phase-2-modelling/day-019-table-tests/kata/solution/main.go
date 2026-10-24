package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var lines, n int
	fmt.Fscan(in, &lines, &n)
	diff := make([]int, lines+2)
	for i := 0; i < n; i++ {
		var l, r int
		fmt.Fscan(in, &l, &r)
		diff[l]++
		diff[r+1]--
	}

	covered := 0
	var gaps []string
	start := 0 // first line of the current gap, 0 when not in one
	run := 0
	for line := 1; line <= lines+1; line++ {
		run += diff[line]
		if line <= lines && run > 0 {
			covered++
		}
		inGap := line <= lines && run == 0
		if inGap && start == 0 {
			start = line
		}
		if !inGap && start != 0 {
			end := line - 1
			if start == end {
				gaps = append(gaps, strconv.Itoa(start))
			} else {
				gaps = append(gaps, strconv.Itoa(start)+"-"+strconv.Itoa(end))
			}
			start = 0
		}
	}

	p := covered * 1000 / lines
	fmt.Fprintf(out, "%d/%d lines (%d.%d%%)\n", covered, lines, p/10, p%10)
	if len(gaps) == 0 {
		fmt.Fprintln(out, "none")
		return
	}
	fmt.Fprintln(out, strings.Join(gaps, " "))
}
