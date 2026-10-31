// Command gen prints the big case for this kata: the full 200 000 attempts from 5 000 phones,
// with 10-digit times and 16-digit phone numbers so the input is about 6 MB. Reading it a byte at
// a time, or printing each answer without a bufio.Writer, means millions of system calls.
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
	phones := make([]string, 5000)
	for i := range phones {
		phones[i] = "07" + strconv.Itoa(10000000000000+r.IntN(90000000000000))
	}
	out.WriteString(strconv.Itoa(n) + "\n")
	t := 900000000
	for i := 0; i < n; i++ {
		t += r.IntN(4)
		result := " OK\n"
		if r.Float64() < 0.6 {
			result = " FAIL\n"
		}
		out.WriteString(strconv.Itoa(t) + " " + phones[r.IntN(len(phones))] + result)
	}
}
