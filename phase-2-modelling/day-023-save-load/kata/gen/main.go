// Command gen prints the big case for this kata: the full 100 000 payments to 20 000 tills,
// about 8 MB of JSON. Fields come in shuffled orders, "reversed" is true, false or missing, and
// some objects are spread over several lines. Decoding it through a reader that fetches one byte
// per system call (no bufio.Reader in front of os.Stdin) takes seconds.
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
	const n, tills = 100000, 20000
	out.WriteString("[\n")
	fields := make([]string, 0, 4)
	for i := 0; i < n; i++ {
		fields = fields[:0]
		fields = append(fields,
			`"till": "`+strconv.Itoa(100000+r.IntN(tills))+`"`,
			`"amount_ksh": `+strconv.Itoa(r.IntN(1000000001)),
			`"phone": "07`+strconv.Itoa(10000000+r.IntN(90000000))+`"`)
		switch x := r.IntN(10); {
		case x < 3:
			fields = append(fields, `"reversed": true`)
		case x < 6:
			fields = append(fields, `"reversed": false`)
		}
		r.Shuffle(len(fields), func(a, b int) { fields[a], fields[b] = fields[b], fields[a] })
		sep := ", "
		if i%7 == 0 {
			sep = ",\n    "
		}
		out.WriteString("  {")
		for j, f := range fields {
			if j > 0 {
				out.WriteString(sep)
			}
			out.WriteString(f)
		}
		out.WriteString("}")
		if i < n-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("]\n")
}
