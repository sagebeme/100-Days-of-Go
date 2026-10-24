package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	titles := make([]string, n)
	for i := range titles {
		fmt.Fscan(in, &titles[i])
	}
	sort.Strings(titles)

	var q int
	fmt.Fscan(in, &q)
	for range q {
		var p string
		fmt.Fscan(in, &p)
		// Every title with prefix p sits in one block of the sorted slice, starting at
		// the first title that is >= p.
		i := sort.SearchStrings(titles, p)
		var picks []string
		for ; i < n && len(picks) < 3 && strings.HasPrefix(titles[i], p); i++ {
			picks = append(picks, titles[i])
		}
		if len(picks) == 0 {
			fmt.Fprintln(out, "-")
		} else {
			fmt.Fprintln(out, strings.Join(picks, " "))
		}
	}
}
