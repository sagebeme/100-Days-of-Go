// Command gen prints the big case for this kata: a 992 × 997 grid of 9s with five corridors of
// 1s running top to bottom, joined at alternate ends, so the only fast route snakes down, up,
// down, up and down again. The walls are 248 blocks thick: cutting through one costs at least
// 9 × 248 minutes, more than the 2 × 991 + 249 of going round, so the snake really is the
// answer. A right-and-down DP can't go up at all, and relaxing every block in reading order
// until nothing changes moves the answer only one block up each pass: about 2000 passes over
// a million blocks.
package main

import (
	"bufio"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const t = 248       // wall thickness
	const corridors = 5 // at columns 0, 249, 498, 747, 996
	const rows = 992    // 8t > 2(rows-1), so a wall never pays to cut
	const cols = 1 + (corridors-1)*(t+1)
	out.WriteString(strconv.Itoa(rows) + " " + strconv.Itoa(cols) + "\n")
	line := make([]byte, cols+1)
	line[cols] = '\n'
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			line[c] = '9'
			k := c / (t + 1) // the corridor at or to the left of c
			switch {
			case c%(t+1) == 0:
				line[c] = '1' // a corridor
			case k%2 == 0 && r == rows-1, k%2 == 1 && r == 0:
				line[c] = '1' // the connector to the next corridor: bottom, then top, ...
			}
		}
		out.Write(line)
	}
}
