// Command gen prints the big case for this kata: the full 200 000 operations, 70% bookings on 500
// trips and 30% cancels of earlier bookings, with 16-character IDs and dated trip names so the
// input is about 7 MB. Reading it a byte at a time, or printing each answer without a
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
	trips := make([]string, 500)
	routes := []string{"46K", "111", "125", "33", "58", "105", "237"}
	for i := range trips {
		hhmm := strconv.Itoa(500 + i%17*100 + i*7%60)
		if len(hhmm) < 4 {
			hhmm = "0" + hhmm
		}
		trips[i] = routes[i%len(routes)] + "-" + hhmm + "-oct" + strconv.Itoa(10+i%20)
	}
	var ids []string
	out.WriteString(strconv.Itoa(n) + "\n")
	for i := 0; i < n; i++ {
		if len(ids) > 0 && r.Float64() < 0.3 {
			out.WriteString("CANCEL " + ids[r.IntN(len(ids))] + "\n")
			continue
		}
		id := "bk" + strconv.Itoa(10000000000000+i*7919%90000000000000)
		ids = append(ids, id)
		out.WriteString("BOOK " + id + " " + trips[r.IntN(len(trips))] + " " + strconv.Itoa(1+r.IntN(33)) + "\n")
	}
}
