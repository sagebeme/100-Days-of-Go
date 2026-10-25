package main

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
)

type standing struct {
	user   string
	solved int
	hints  int
}

type solve struct {
	user string
	day  int
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q int
	fmt.Fscan(in, &q)
	done := map[solve]bool{}
	index := map[string]int{} // user -> position in board
	var board []standing
	for range q {
		var cmd string
		fmt.Fscan(in, &cmd)
		if cmd == "SOLVE" {
			var s solve
			var hints int
			fmt.Fscan(in, &s.user, &s.day, &hints)
			if done[s] {
				continue // only the first solve of a day counts
			}
			done[s] = true
			i, ok := index[s.user]
			if !ok {
				i = len(board)
				index[s.user] = i
				board = append(board, standing{user: s.user})
			}
			board[i].solved++
			board[i].hints += hints
			continue
		}
		var k int
		fmt.Fscan(in, &k)
		ranked := slices.Clone(board)
		slices.SortFunc(ranked, func(a, b standing) int {
			return cmp.Or(
				cmp.Compare(b.solved, a.solved), // more solved first
				cmp.Compare(a.hints, b.hints),   // then fewer hints
				strings.Compare(a.user, b.user), // then by name
			)
		})
		if len(ranked) == 0 {
			fmt.Fprintln(out, "-")
			continue
		}
		parts := make([]string, 0, k)
		for _, s := range ranked[:min(k, len(ranked))] {
			parts = append(parts, fmt.Sprintf("%s:%d:%d", s.user, s.solved, s.hints))
		}
		fmt.Fprintln(out, strings.Join(parts, " "))
	}
}
