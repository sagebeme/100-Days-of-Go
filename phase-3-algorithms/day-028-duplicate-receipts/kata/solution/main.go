package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	var goals string
	fmt.Fscan(in, &n, &goals)

	// diff is Gor goals minus Leopards goals so far. A stretch is a draw when the
	// diff at its end equals the diff just before its start. Count equal diffs.
	seen := map[int]int{0: 1} // before any goal, the diff is 0
	diff, draws := 0, 0
	for i := 0; i < len(goals); i++ {
		if goals[i] == 'G' {
			diff++
		} else {
			diff--
		}
		draws += seen[diff]
		seen[diff]++
	}
	fmt.Println(draws)
}
