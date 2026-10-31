// Command gen prints the big case for this kata: the full 200 000 events over 20 000 IDs with
// W = 30 000, 10-digit times and every ID the full 20 characters, so the input is as big as the
// limits allow (about 6 MB). Reading it a byte at a time, or printing each verdict without a
// bufio.Writer, means millions of system calls.
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
	const n = 200000
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ids := make([]string, 20000)
	for i := range ids {
		b := []byte("evt_")
		for len(b) < 20 {
			b = append(b, chars[r.IntN(len(chars))])
		}
		ids[i] = string(b)
	}
	out.WriteString(strconv.Itoa(n) + " 30000\n")
	t := 999000000
	for i := 0; i < n; i++ {
		t += r.IntN(3)
		out.WriteString(strconv.Itoa(t) + " " + ids[r.IntN(len(ids))] + "\n")
	}
}
