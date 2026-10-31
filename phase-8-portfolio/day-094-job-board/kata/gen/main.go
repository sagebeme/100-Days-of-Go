// Command gen prints the big case for this kata: the full 100 000 titles (random words using the
// letters a to y) and 100 000 prefixes. Two thirds of the prefixes start with `z` or `yyyy`, which
// sort at or past the end of the list, so scanning the sorted titles from the start for each
// prefix is close to 100 000 × 100 000 steps.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n, q = 100000, 100000
	seen := map[string]bool{}
	out.WriteString(strconv.Itoa(n) + "\n")
	for len(seen) < n {
		b := make([]byte, 5+r.IntN(16))
		for i := range b {
			b[i] = byte('a' + r.IntN(25))
		}
		if s := string(b); !seen[s] {
			seen[s] = true
			out.WriteString(s + "\n")
		}
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		switch r.IntN(3) {
		case 0:
			out.WriteString("z" + strings.Repeat("a", r.IntN(11)) + "\n")
		case 1:
			out.WriteString(string(rune('a'+r.IntN(6))) + string(rune('a'+r.IntN(3))) + "\n")
		default:
			out.WriteString("yyyy\n")
		}
	}
}
