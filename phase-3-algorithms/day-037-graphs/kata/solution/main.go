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
	start := 0
	for r := range grid {
		fmt.Fscan(in, &grid[r])
		for c := 0; c < cols; c++ {
			if grid[r][c] == 'S' {
				start = r*cols + c
			}
		}
	}

	// BFS over cells numbered r*cols + c. dist -1 means not reached yet.
	dist := make([]int, rows*cols)
	for i := range dist {
		dist[i] = -1
	}
	dist[start] = 0
	queue := []int{start}
	dr := [4]int{-1, 1, 0, 0}
	dc := [4]int{0, 0, -1, 1}
	for head := 0; head < len(queue); head++ {
		cell := queue[head]
		r, c := cell/cols, cell%cols
		if grid[r][c] == 'T' {
			fmt.Println(dist[cell])
			return
		}
		for k := range 4 {
			nr, nc := r+dr[k], c+dc[k]
			if nr < 0 || nr >= rows || nc < 0 || nc >= cols || grid[nr][nc] == '#' {
				continue
			}
			if next := nr*cols + nc; dist[next] < 0 {
				dist[next] = dist[cell] + 1
				queue = append(queue, next)
			}
		}
	}
	fmt.Println(-1)
}
