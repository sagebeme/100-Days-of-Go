// Command gen prints the big case for this kata: 100 budgets and 160 000 transactions over about
// two years, 150 categories (a third with no budget), about 4.5 MB in all. Budgets range
// from 0 to more than a month's spending, so some categories alert on day one, some late in the
// month and some never. Reading it one byte per system call (fmt.Fscan straight on os.Stdin,
// with no bufio.Reader) takes seconds.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
	"time"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const b, n, ncat = 100, 160000, 150
	cats := make([]string, ncat)
	seen := map[string]bool{}
	for i := range cats {
		for cats[i] == "" || seen[cats[i]] {
			w := make([]byte, 8+r.IntN(5))
			for j := range w {
				w[j] = byte('a' + r.IntN(26))
			}
			cats[i] = string(w)
		}
		seen[cats[i]] = true
	}
	// About 230 transactions a day and 150 categories: each category sees roughly 45 a month,
	// worth about 22 000 000 shillings.
	out.WriteString(strconv.Itoa(b) + "\n")
	for _, c := range cats[:b] {
		out.WriteString(c + " " + strconv.Itoa(r.IntN(30000001)) + "\n")
	}
	out.WriteString(strconv.Itoa(n) + "\n")
	d := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		if i > 0 && i%230 == 0 {
			d = d.AddDate(0, 0, 1)
		}
		out.WriteString(d.Format("2006-01-02") + " " + cats[r.IntN(ncat)] + " " + strconv.Itoa(r.IntN(1000001)) + "\n")
	}
}
