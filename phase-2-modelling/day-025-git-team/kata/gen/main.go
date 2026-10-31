// Command gen prints the big case for this kata: the full 100 000 headers, most of them close to
// the 100-character limit, about 8 MB in all. Every "!" header is invalid in some small way (a
// capital letter, a bad scope, a missing or doubled space), so only a careful check finds that
// the answer is a minor release. Reading it one byte per system call (a bufio.Scanner straight on
// a reader that fetches a byte at a time) takes seconds.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

var valid = []string{"fix: ", "docs: ", "chore(deps): ", "perf: ", "refactor(day-24): ", "test(x-1): ", "fix(ledger): "}

var trap = []string{
	"Feat!: ", "feat(Day-24)!: ", "feat(a b)!: ", "feat!:", "feat!:  ", "feat !: ", "fix(api)! : ",
	"feat()!: ", "feat(x)(y)!: ", "feat!!: ", "feat-x!: ", "(api)!: ", "!: ", "feat!;  ", "feat_x!: ",
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const n = 100000
	out.WriteString("v41.999999.1000000\n" + strconv.Itoa(n) + "\n")
	const letters = "abcdefghijklmnopqrstuvwxyz ABC.,-()!:0123456789"
	for i := 0; i < n; i++ {
		var head string
		switch x := r.IntN(10); {
		case x < 5:
			head = trap[r.IntN(len(trap))]
		case i%5000 == 4321: // a few real features
			head = "feat(export): "
		default:
			head = valid[r.IntN(len(valid))]
		}
		var sb strings.Builder
		sb.WriteString(head)
		sb.WriteByte('a' + byte(r.IntN(26)))
		for total := 60 + r.IntN(41); sb.Len() < total; {
			sb.WriteByte(letters[r.IntN(len(letters))])
		}
		out.WriteString(sb.String() + "\n")
	}
}
