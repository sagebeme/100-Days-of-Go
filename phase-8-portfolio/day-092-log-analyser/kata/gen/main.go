// Command gen prints the big case for this kata: the full 200 000 requests, all inside ten minutes
// of a morning spike (09:00:00 to 09:10:00), sorted. Counting every window from scratch is
// 200 000 × 200 000 / 2 steps.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
)

func two(x int) string {
	if x < 10 {
		return "0" + strconv.Itoa(x)
	}
	return strconv.Itoa(x)
}

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n = 200000
	ts := make([]int, n)
	for i := range ts {
		ts[i] = 9*3600 + r.IntN(601)
	}
	slices.Sort(ts)
	out.WriteString(strconv.Itoa(n) + "\n")
	for _, t := range ts {
		out.WriteString(two(t/3600) + ":" + two(t/60%60) + ":" + two(t%60) + "\n")
	}
}
