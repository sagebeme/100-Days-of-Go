package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
)

// item is a cell waiting in the priority queue with its best known time.
type item struct{ time, cell int }

type queue []item

func (q queue) Len() int           { return len(q) }
func (q queue) Less(i, j int) bool { return q[i].time < q[j].time }
func (q queue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *queue) Push(x any)        { *q = append(*q, x.(item)) }
func (q *queue) Pop() any {
	old := *q
	x := old[len(old)-1]
	*q = old[:len(old)-1]
	return x
}

func main() {
	in := bufio.NewReader(os.Stdin)
	var rows, cols int
	fmt.Fscan(in, &rows, &cols)
	grid := make([]string, rows)
	for r := range grid {
		fmt.Fscan(in, &grid[r])
	}

	// Dijkstra: always settle the unsettled cell with the smallest time next.
	// Every cost is positive, so once a cell is popped its time is final.
	const unknown = -1
	best := make([]int, rows*cols)
	for i := range best {
		best[i] = unknown
	}
	best[0] = 0
	q := &queue{{0, 0}}
	target := rows*cols - 1
	dr := [4]int{-1, 1, 0, 0}
	dc := [4]int{0, 0, -1, 1}
	for q.Len() > 0 {
		cur := heap.Pop(q).(item)
		if cur.time > best[cur.cell] {
			continue // a stale entry: we already found a faster way here
		}
		if cur.cell == target {
			break
		}
		r, c := cur.cell/cols, cur.cell%cols
		for k := range 4 {
			nr, nc := r+dr[k], c+dc[k]
			if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
				continue
			}
			next := nr*cols + nc
			t := cur.time + int(grid[nr][nc]-'0')
			if best[next] == unknown || t < best[next] {
				best[next] = t
				heap.Push(q, item{t, next})
			}
		}
	}
	fmt.Println(best[target])
}
