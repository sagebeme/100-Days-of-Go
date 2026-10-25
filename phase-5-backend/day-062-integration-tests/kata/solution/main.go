package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

type run struct{ commit, test string }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	// results[run] has bit 1 set if it ever passed and bit 2 if it ever failed.
	results := map[run]int{}
	for i := 0; i < n; i++ {
		var commit, test, result string
		fmt.Fscan(in, &commit, &test, &result)
		bit := 1
		if result == "FAIL" {
			bit = 2
		}
		results[run{commit, test}] |= bit
	}

	flaky := map[string]int{}
	for r, bits := range results {
		if bits == 3 {
			flaky[r.test]++
		}
	}
	if len(flaky) == 0 {
		fmt.Fprintln(out, "NONE")
		return
	}
	names := make([]string, 0, len(flaky))
	for name := range flaky {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if flaky[names[i]] != flaky[names[j]] {
			return flaky[names[i]] > flaky[names[j]]
		}
		return names[i] < names[j]
	})
	for _, name := range names {
		fmt.Fprintln(out, name, flaky[name])
	}
}
