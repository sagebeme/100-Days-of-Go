// Command gen prints the big case for this kata: the full 200 000 results over 500 commits (12-digit
// hashes) and 1 500 tests with long names, about 8 MB of input. Reading it a byte at a time, or
// printing without a bufio.Writer, means millions of system calls.
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
	commits := make([]string, 500)
	for i := range commits {
		h := make([]byte, 12)
		for j := range h {
			h[j] = "0123456789abcdef"[r.IntN(16)]
		}
		commits[i] = string(h)
	}
	kinds := []string{"TestBookingJourney", "TestLastSeatRace", "TestHealthCheck", "TestRefundFlow", "TestPaymentWebhook"}
	tests := make([]string, 1500)
	for i := range tests {
		tests[i] = kinds[i%len(kinds)] + "_" + strconv.Itoa(i)
	}
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		result := " PASS\n"
		if r.Float64() >= 0.7 {
			result = " FAIL\n"
		}
		out.WriteString(commits[r.IntN(len(commits))] + " " + tests[r.IntN(len(tests))] + result)
	}
}
