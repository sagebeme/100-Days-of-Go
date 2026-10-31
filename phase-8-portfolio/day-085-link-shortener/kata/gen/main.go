// Command gen prints the big case for this kata: the full 100 000 URLs (60 to 100 characters, a
// quarter of them repeats) and 100 000 lookups, half of them real codes so the answer prints the
// long URL back. Input and output come to about 14 MB, so reading a byte at a time, or printing
// each line without a bufio.Writer, means millions of system calls.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 11)
	for n > 0 {
		buf = append(buf, alphabet[n%62])
		n /= 62
	}
	slices.Reverse(buf)
	return string(buf)
}

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n, q, start = 100000, 100000, 1000000000000
	sites := []string{"https://duka.co.ke/products/", "https://kiungo.ke/blog/2026/", "https://mtoto.school/fees/term-", "https://twende.co.ke/trips/"}
	words := []string{"sufuria", "unga", "matatu", "nairobi", "mombasa", "kisumu", "booking", "seat", "receipt", "offer"}
	var urls, codes []string
	out.WriteString(strconv.Itoa(n) + " " + strconv.Itoa(start) + "\n")
	for i := 0; i < n; i++ {
		if len(urls) > 0 && r.IntN(4) == 0 {
			out.WriteString(urls[r.IntN(len(urls))] + "\n")
			continue
		}
		u := sites[r.IntN(len(sites))] + strconv.Itoa(i)
		for want := 60 + r.IntN(41); len(u) < want; {
			u += "-" + words[r.IntN(len(words))]
		}
		u = u[:min(len(u), 100)]
		codes = append(codes, encode(uint64(start+len(urls))))
		urls = append(urls, u)
		out.WriteString(u + "\n")
	}
	out.WriteString(strconv.Itoa(q) + "\n")
	for i := 0; i < q; i++ {
		c := codes[r.IntN(len(codes))]
		switch r.IntN(6) {
		case 0:
			c = "0" + c // same number, different code
		case 1:
			c = c[:len(c)-1] + "!"
		case 2:
			b := make([]byte, 1+r.IntN(12))
			for j := range b {
				b[j] = alphabet[r.IntN(62)]
			}
			c = string(b)
		}
		out.WriteString(c + "\n")
	}
}
