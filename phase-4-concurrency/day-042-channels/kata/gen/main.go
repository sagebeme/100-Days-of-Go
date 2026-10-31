// Command gen prints the big case for this kata: the full 100 000 customers with q = 1 and the
// minutes adding up to the 1 000 000 maximum, so the agent takes a million turns. A queue that
// drops its front by copying everything left one place (append(q[:0], q[1:]...)) does up to
// 100 000 steps on each of those turns.
package main

import (
	"bufio"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	r := rand.New(rand.NewPCG(1, 2))
	const n, q, total = 100000, 1, 1000000
	minutes := make([]int, n)
	sum := 0
	for i := range minutes {
		minutes[i] = 1 + r.IntN(17) // 9 on average
		sum += minutes[i]
	}
	// Give what's left of the million to a few long jobs (each at most 50 000 minutes).
	for i := 0; sum < total; i = (i + 9973) % n {
		add := min(total-sum, 50000-minutes[i], 20000)
		minutes[i] += add
		sum += add
	}
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out.WriteString(strconv.Itoa(n) + " " + strconv.Itoa(q) + "\n")
	for i, m := range minutes {
		// A unique name: a random letter prefix, then the customer's number.
		name := []byte{chars[r.IntN(52)], chars[r.IntN(len(chars))], chars[r.IntN(len(chars))]}
		name = strconv.AppendInt(name, int64(i), 10)
		out.Write(name)
		out.WriteString(" " + strconv.Itoa(m) + "\n")
	}
}
