package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var n int
	fmt.Fscan(in, &n)
	ts := make([]int, n)
	for i := range ts {
		var s string
		fmt.Fscan(in, &s)
		var h, m, sec int
		fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
		ts[i] = h*3600 + m*60 + sec
	}

	best, start := 0, 0
	j := 0 // the last request inside the window that starts at ts[i]
	for i := range ts {
		j = max(j, i)
		for j+1 < n && ts[j+1] <= ts[i]+59 {
			j++
		}
		if c := j - i + 1; c > best {
			best, start = c, ts[i]
		}
	}
	fmt.Printf("%d %02d:%02d:%02d\n", best, start/3600, start/60%60, start%60)
}
