package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

type stats struct {
	name                string
	earnings, km, trips int
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n, k int
	fmt.Fscan(in, &n, &k)
	byName := map[string]*stats{}
	for i := 0; i < n; i++ {
		var name string
		var km, fare int
		fmt.Fscan(in, &name, &km, &fare)
		s, ok := byName[name]
		if !ok {
			s = &stats{name: name}
			byName[name] = s
		}
		s.earnings += fare
		s.km += km
		s.trips++
	}

	list := make([]stats, 0, len(byName))
	for _, s := range byName {
		list = append(list, *s)
	}
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.earnings != b.earnings {
			return a.earnings > b.earnings
		}
		if a.km != b.km {
			return a.km > b.km
		}
		return a.name < b.name
	})
	for i := 0; i < k && i < len(list); i++ {
		s := list[i]
		fmt.Fprintln(out, s.name, s.earnings, s.km, s.trips)
	}
}
