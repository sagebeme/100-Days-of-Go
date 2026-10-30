package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

// place finds the first run of k free seats, marks it taken and returns its row and column
// (from 1). ok is false if no row has room.
func place(seats [][]byte, k int) (row, col int, ok bool) {
	for r, line := range seats {
		run := 0
		for c, s := range line {
			if s != '.' {
				run = 0
				continue
			}
			run++
			if run == k {
				start := c - k + 1
				for i := start; i <= c; i++ {
					line[i] = 'X'
				}
				return r + 1, start + 1, true
			}
		}
	}
	return 0, 0, false
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var rows, cols, n int
	fmt.Fscan(in, &rows, &cols, &n)
	seats := make([][]byte, rows)
	for i := range seats {
		var line string
		fmt.Fscan(in, &line)
		seats[i] = []byte(line)
	}
	for i := 0; i < n; i++ {
		var k int
		fmt.Fscan(in, &k)
		if r, c, ok := place(seats, k); ok {
			fmt.Fprintln(out, r, c)
		} else {
			fmt.Fprintln(out, "SOLD OUT")
		}
	}
	free := 0
	for _, line := range seats {
		free += bytes.Count(line, []byte("."))
	}
	fmt.Fprintf(out, "free %d\n", free)
}
