package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// parse turns "a.b.c.d" into a 32-bit number. Each part must be 1 to 3 digits, at
// most 255, with no leading zero.
func parse(s string) (uint32, bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return 0, false
	}
	var v uint32
	for _, p := range parts {
		if len(p) == 0 || len(p) > 3 || (len(p) > 1 && p[0] == '0') {
			return 0, false
		}
		for i := 0; i < len(p); i++ {
			if p[i] < '0' || p[i] > '9' {
				return 0, false
			}
		}
		n, _ := strconv.Atoi(p)
		if n > 255 {
			return 0, false
		}
		v = v<<8 | uint32(n)
	}
	return v, true
}

type block struct{ network, mask uint32 }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var m int
	fmt.Fscan(in, &m)
	blocks := make([]block, m)
	for i := range blocks {
		var cidr string
		fmt.Fscan(in, &cidr)
		ip, bits, _ := strings.Cut(cidr, "/")
		v, _ := parse(ip)
		b, _ := strconv.Atoi(bits)
		var mask uint32
		if b > 0 {
			mask = ^uint32(0) << (32 - b)
		}
		blocks[i] = block{v & mask, mask}
	}

	var q int
	fmt.Fscan(in, &q)
	for i := 0; i < q; i++ {
		var s string
		fmt.Fscan(in, &s)
		v, ok := parse(s)
		if !ok {
			fmt.Fprintln(out, "INVALID")
			continue
		}
		verdict := "ALLOW"
		for _, b := range blocks {
			if v&b.mask == b.network {
				verdict = "BLOCK"
				break
			}
		}
		fmt.Fprintln(out, verdict)
	}
}
