// Command gen prints the big case for this kata: the full 100 000 lines, each padded with spaces to
// about 40 characters, the most the limits allow. Most are long valid amounts (up to 13 digits
// with commas, so the answers are long too) and the rest are the README's tricky cases. Input and
// output come to about 6 MB, so reading a byte at a time, or printing each answer without a
// bufio.Writer, means millions of system calls.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

// commas writes n with a comma every three digits.
func commas(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const q = 100000
	tricky := []string{"Ksh 1,200.5", "kes99", "-ksh 20.05", "1234567", "1,23", ".5", "5.", "--1", "USD 5",
		"KES -1,234,567,890,123.45", "", "12.345", "0001,000", "-0", "1 000", "Ksh--5", "-Ksh -5", "12,345,6789"}
	signs := []string{"", "-"}
	currencies := []string{"", "Ksh ", "KES", "kSh  ", "kes   "}
	cents := []string{"", ".5", ".05", ".99"}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		var s string
		if r.IntN(4) == 0 {
			s = tricky[r.IntN(len(tricky))]
		} else {
			n := r.Int64N(10000000000000)
			body := strconv.FormatInt(n, 10)
			if r.IntN(3) > 0 {
				body = commas(n)
			}
			s = signs[r.IntN(2)] + currencies[r.IntN(len(currencies))] + body + cents[r.IntN(len(cents))]
		}
		// Pad both ends with spaces, up to 40 characters in all.
		pad := 40 - len(s)
		left := r.IntN(pad + 1)
		out.WriteString(strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left) + "\n")
	}
}
