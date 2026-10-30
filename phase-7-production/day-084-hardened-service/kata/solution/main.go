package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
)

type stats struct {
	latencies []int
	errors    int
}

// p95 returns the nearest-rank 95th percentile, sorting l in place.
func p95(l []int) int {
	slices.Sort(l)
	return l[(95*len(l)+99)/100-1]
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	routes := map[string]*stats{}
	all := &stats{}
	perMinute := map[int]int{}
	for range n {
		var ts, status, ms int
		var route string
		fmt.Fscan(in, &ts, &route, &status, &ms)
		s, ok := routes[route]
		if !ok {
			s = &stats{}
			routes[route] = s
		}
		s.latencies = append(s.latencies, ms)
		all.latencies = append(all.latencies, ms)
		if status >= 500 {
			s.errors++
			all.errors++
			perMinute[ts/60]++
		}
	}

	names := make([]string, 0, len(routes))
	for name := range routes {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		s := routes[name]
		fmt.Fprintln(out, name, len(s.latencies), s.errors, p95(s.latencies))
	}
	fmt.Fprintln(out, "ALL", len(all.latencies), all.errors, p95(all.latencies))

	worst, most := -1, 0
	for m, e := range perMinute {
		if e > most || (e == most && m < worst) {
			worst, most = m, e
		}
	}
	if worst < 0 {
		fmt.Fprintln(out, "worst minute none")
	} else {
		fmt.Fprintf(out, "worst minute %d with %d errors\n", worst, most)
	}
}
