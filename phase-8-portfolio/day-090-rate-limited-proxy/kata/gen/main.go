// Command gen prints the big case for this kata: the full 200 000 requests with b = 5 and r = 1,
// 12-digit times and long IPv6 client names, about 8 MB of input. Twenty busy clients send most
// of the traffic, faster than their buckets refill, so there are plenty of denials. Reading it a
// byte at a time, or printing each verdict without a bufio.Writer, means millions of system calls.
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
	hex := func() string { return strconv.FormatUint(uint64(0x1000+r.IntN(0xf000)), 16) }
	clients := make([]string, 5000)
	for i := range clients {
		clients[i] = "2c0f:fe38:" + hex() + ":" + hex() + ":" + hex() + ":" + hex() + "::1"
	}
	out.WriteString("5 1 " + strconv.Itoa(n) + "\n")
	t := 999000000000
	for i := 0; i < n; i++ {
		t += r.IntN(21)
		c := clients[r.IntN(len(clients))]
		if r.Float64() < 0.7 {
			c = clients[r.IntN(20)]
		}
		out.WriteString(strconv.Itoa(t) + " " + c + "\n")
	}
}
