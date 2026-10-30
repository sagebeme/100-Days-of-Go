package main

import (
	"bufio"
	"fmt"
	"os"
)

type ride struct{ from, to, fare int }

func main() {
	in := bufio.NewReader(os.Stdin)
	var n, m, k int
	fmt.Fscan(in, &n, &m, &k)
	rides := make([]ride, m)
	for i := range rides {
		fmt.Fscan(in, &rides[i].from, &rides[i].to, &rides[i].fare)
	}
	var s, d int
	fmt.Fscan(in, &s, &d)

	// Bellman-Ford, stopped after K+1 rounds. After round r, cost[v] is the cheapest
	// fare to v using at most r rides. Each round reads last round's costs (prev), so
	// one round can't chain two rides together.
	const inf = 1 << 60
	cost := make([]int, n+1)
	for i := range cost {
		cost[i] = inf
	}
	cost[s] = 0
	for range k + 1 {
		prev := append([]int(nil), cost...)
		for _, r := range rides {
			if prev[r.from] != inf && prev[r.from]+r.fare < cost[r.to] {
				cost[r.to] = prev[r.from] + r.fare
			}
		}
	}
	if cost[d] == inf {
		fmt.Println(-1)
	} else {
		fmt.Println(cost[d])
	}
}
