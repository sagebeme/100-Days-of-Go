// Command gen prints the big case for this kata: the full 100 000 pairs, a few megabytes of
// versions. Many pairs are equal once leading zeros and missing ".0" parts are ignored, and
// many differ in a single late part, so every pair has to be compared part by part. Reading it
// one byte per system call (fmt.Fscan straight on os.Stdin, with no bufio.Reader) takes seconds.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

// part writes x with up to `pad` leading zeros, never more than 9 digits in all.
func part(sb *strings.Builder, r *rand.Rand, x, pad int) {
	s := strconv.Itoa(x)
	if z := min(r.IntN(pad+1), 9-len(s)); z > 0 {
		sb.WriteString(strings.Repeat("0", z))
	}
	sb.WriteString(s)
}

func write(sb *strings.Builder, r *rand.Rand, parts []int, pad int) {
	for i, x := range parts {
		if i > 0 {
			sb.WriteByte('.')
		}
		part(sb, r, x, pad)
	}
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const q = 100000
	out.WriteString(strconv.Itoa(q) + "\n")
	var sb strings.Builder
	for i := 0; i < q; i++ {
		k := 3 + r.IntN(6)
		a := make([]int, k)
		for j := range a {
			a[j] = r.IntN(1000)
		}
		b := append([]int(nil), a...)
		switch x := r.IntN(10); {
		case x < 4: // the same version, written differently
			for len(b) < 10 && r.IntN(2) == 0 {
				b = append(b, 0)
			}
		case x < 8: // one part (usually a late one) is one higher or lower
			j := k - 1 - r.IntN(2)
			if b[j] == 0 || r.IntN(2) == 0 {
				b[j]++
			} else {
				b[j]--
			}
		default: // an extra non-zero part on the end, or a long random part
			if len(b) < 10 && r.IntN(2) == 0 {
				b = append(b, 1+r.IntN(999999999))
			} else {
				b[r.IntN(k)] = r.IntN(1000000000)
			}
		}
		if r.IntN(2) == 0 {
			a, b = b, a
		}
		sb.Reset()
		write(&sb, r, a, 3)
		sb.WriteByte(' ')
		write(&sb, r, b, 3)
		sb.WriteByte('\n')
		out.WriteString(sb.String())
	}
}
