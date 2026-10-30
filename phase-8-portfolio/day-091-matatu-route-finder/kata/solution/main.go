package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"slices"
)

type edge struct{ u, v, w int }

func main() {
	in := bufio.NewReader(os.Stdin)

	var n, m, k, s, d int
	fmt.Fscan(in, &n, &m, &k, &s, &d)
	edges := make([]edge, m)
	for i := range edges {
		fmt.Fscan(in, &edges[i].u, &edges[i].v, &edges[i].w)
	}

	const inf = math.MaxInt
	dist := make([]int, n)
	for i := range dist {
		dist[i] = inf
	}
	dist[s] = 0
	for range k + 1 {
		next := slices.Clone(dist)
		for _, e := range edges {
			if dist[e.u] != inf && dist[e.u]+e.w < next[e.v] {
				next[e.v] = dist[e.u] + e.w
			}
		}
		dist = next
	}
	if dist[d] == inf {
		fmt.Println(-1)
		return
	}
	fmt.Println(dist[d])
}
