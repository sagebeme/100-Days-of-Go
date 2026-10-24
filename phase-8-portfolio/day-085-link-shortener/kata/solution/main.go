package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
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
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	var next uint64
	fmt.Fscan(in, &n, &next)
	codeOf := make(map[string]string, n)
	urlOf := make(map[string]string, n)
	for range n {
		var u string
		fmt.Fscan(in, &u)
		c, ok := codeOf[u]
		if !ok {
			c = encode(next)
			next++
			codeOf[u] = c
			urlOf[c] = u
		}
		fmt.Fprintln(out, c)
	}

	var q int
	fmt.Fscan(in, &q)
	for range q {
		var c string
		fmt.Fscan(in, &c)
		if u, ok := urlOf[c]; ok {
			fmt.Fprintln(out, u)
		} else {
			fmt.Fprintln(out, "404")
		}
	}
}
