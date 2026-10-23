package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var rows, cols int
	fmt.Fscan(in, &rows, &cols)
	grid := make([]string, rows)
	r, c := 0, 0
	for i := range grid {
		fmt.Fscan(in, &grid[i])
		for j := 0; j < cols; j++ {
			if grid[i][j] == 'S' {
				r, c = i, j
			}
		}
	}
	var moves string
	fmt.Fscan(in, &moves)

	dr := map[byte]int{'N': -1, 'S': 1}
	dc := map[byte]int{'E': 1, 'W': -1}
	blocked := 0
	for i := 0; i < len(moves); i++ {
		nr, nc := r+dr[moves[i]], c+dc[moves[i]]
		if nr < 0 || nr >= rows || nc < 0 || nc >= cols || grid[nr][nc] == '#' {
			blocked++
			continue
		}
		r, c = nr, nc
	}
	fmt.Println(r+1, c+1)
	fmt.Println(blocked)
}
