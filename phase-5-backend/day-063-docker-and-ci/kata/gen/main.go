// Command gen prints the big case for this kata: the full 100 000 jobs (random 8-letter names)
// and 200 000 needs, all pointing forward along a hidden random order so there's no cycle.
// Scanning every job to find the smallest ready one, once per job run, is 100 000 × 100 000 steps.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n, m = 100000, 200000
	names := make([]string, 0, n)
	seen := map[string]bool{}
	for len(names) < n {
		b := make([]byte, 8)
		for i := range b {
			b[i] = byte('a' + r.IntN(26))
		}
		if s := string(b); !seen[s] {
			seen[s] = true
			names = append(names, s)
		}
	}
	order := append([]string(nil), names...)
	r.Shuffle(n, func(i, j int) { order[i], order[j] = order[j], order[i] })

	out.WriteString(strconv.Itoa(n) + " " + strconv.Itoa(m) + "\n")
	for i, s := range names {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(s)
	}
	out.WriteByte('\n')
	// A need always points from a later job in order to an earlier one.
	edges := map[[2]int]bool{}
	for len(edges) < m {
		i, j := r.IntN(n), r.IntN(n)
		if i == j {
			continue
		}
		if i > j {
			i, j = j, i
		}
		e := [2]int{j, i}
		if edges[e] {
			continue
		}
		edges[e] = true
		out.WriteString(order[j] + " " + order[i] + "\n")
	}
}
