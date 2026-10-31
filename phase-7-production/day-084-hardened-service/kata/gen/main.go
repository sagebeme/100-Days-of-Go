// Command gen prints the big case for this kata: the full 100 000 log lines over 300 routes, each
// route name 30 characters long and every timestamp 10 digits, so the input is close to the
// biggest the limits allow (about 5 MB). Reading it a byte at a time means millions of system
// calls.
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
	const n = 100000
	kinds := []string{"/api/v1/bookings/seats/", "/api/v1/trips/search/", "/api/v1/mpesa/callback/"}
	routes := make([]string, 300)
	for i := range routes {
		// The kind, then i padded with zeros to exactly 30 characters.
		num := strconv.Itoa(i)
		for len(kinds[i%len(kinds)])+len(num) < 30 {
			num = "0" + num
		}
		routes[i] = kinds[i%len(kinds)] + num
	}
	statuses := []string{"200", "201", "404", "500", "503"}
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		// A weekend of logs, a million seconds that end at 10^9, so busy minutes repeat.
		ts := 999000000 + r.IntN(1000001)
		out.WriteString(strconv.Itoa(ts) + " " + routes[r.IntN(len(routes))] + " " +
			statuses[r.IntN(len(statuses))] + " " + strconv.Itoa(r.IntN(1000001)) + "\n")
	}
}
