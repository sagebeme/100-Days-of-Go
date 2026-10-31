// Command gen prints the big case for this kata: 1000 accounts with 20-character names and
// 100 000 transfers that hit every rule (bad amounts, unknown and same accounts, short funds and
// plenty of OKs), so the input is several megabytes. Reading it one byte per system call
// (fmt.Fscan straight on os.Stdin, with no bufio.Reader) takes seconds.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func name(r *rand.Rand) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 20)
	for j := range b {
		b[j] = chars[r.IntN(len(chars))]
	}
	return string(b)
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const n, q = 1000, 100000
	names := make([]string, n)
	seen := map[string]bool{}
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := range names {
		for names[i] == "" || seen[names[i]] {
			names[i] = name(r)
		}
		seen[names[i]] = true
		out.WriteString(names[i] + " " + strconv.Itoa(r.IntN(1000000001)) + "\n")
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		from, to := names[r.IntN(n)], names[r.IntN(n)]
		amount := 1 + r.IntN(1000000000)
		switch x := r.IntN(100); {
		case x < 3:
			amount = -r.IntN(1000000001)
		case x < 6:
			to = name(r) // almost surely not an account
		case x < 9:
			to = from
		case x < 70:
			amount = 1 + r.IntN(5000000)
		}
		out.WriteString("QFT" + strconv.Itoa(i) + " " + from + " " + to + " " + strconv.Itoa(amount) + "\n")
	}
}
