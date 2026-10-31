// Command gen prints the big case for this kata: the full 200 blocks and 100 000 addresses. Most
// are real addresses, some have octal-style leading zeros, and 35% are long hostnames people pasted
// instead of an address, so the input is about 8 MB. Reading it a byte at a time, or
// printing each answer without a bufio.Writer, means millions of system calls.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const m, q = 200, 100000
	octet := func() string { return strconv.Itoa(r.IntN(256)) }
	addr := func() string { return octet() + "." + octet() + "." + octet() + "." + octet() }
	out.WriteString(strconv.Itoa(m) + "\n")
	for i := 0; i < m; i++ {
		out.WriteString(addr() + "/" + strconv.Itoa(8+r.IntN(25)) + "\n")
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	host := make([]byte, 0, 256)
	for i := 0; i < q; i++ {
		x := r.Float64()
		switch {
		case x < 0.1:
			p := strconv.Itoa(r.IntN(100))
			if len(p) == 1 {
				p = "0" + p
			}
			out.WriteString(strconv.Itoa(r.IntN(300)) + "." + p + ".1.1\n")
		case x < 0.45:
			host = host[:0]
			for n := 150 + r.IntN(100); len(host) < n; {
				if len(host) > 0 && r.IntN(12) == 0 {
					host = append(host, '.')
				}
				host = append(host, byte('a'+r.IntN(26)))
			}
			out.Write(host)
			out.WriteByte('\n')
		default:
			out.WriteString(addr() + "\n")
		}
	}
}
