// Command gen prints the big case for this kata: 70 000 tokens, a mix of every verdict, about 7 MB
// of input. Reading it a byte at a time, or printing each verdict without a bufio.Writer,
// means millions of system calls. (The full 100 000 would push the reference near the limit.)
package main

import (
	"bufio"
	"encoding/base64"
	"math/rand/v2"
	"os"
	"strconv"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	const n = 70000
	enc := base64.RawURLEncoding
	b := func(s string) string { return enc.EncodeToString([]byte(s)) }
	headers := []string{
		b(`{"alg":"HS256","typ":"JWT"}`),
		b(`{"alg":"HS256","typ":"JWT"}`),
		b(`{"alg":"none","typ":"JWT"}`),
	}
	out.WriteString(strconv.Itoa(n) + " 30\n")
	for i := 0; i < n; i++ {
		h := headers[r.IntN(len(headers))]
		sub := "07" + strconv.Itoa(10000000+r.IntN(90000000))
		p := `{"sub":"` + sub + `","iat":` + strconv.Itoa(r.IntN(2001))
		exp := strconv.Itoa(r.IntN(3001))
		if r.Float64() < 0.05 {
			exp = `"12"`
		}
		p += `,"exp":` + exp
		if r.Float64() < 0.3 {
			p += `,"nbf":` + strconv.Itoa(r.IntN(2001))
		}
		p += "}"
		tok := h + "." + b(p) + ".c2ln"
		if r.Float64() < 0.05 {
			tok = "bad.token"
		}
		out.WriteString(strconv.Itoa(r.IntN(3001)) + " " + tok + "\n")
	}
}
