package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const redacted = "[REDACTED]"

var sensitive = []string{"password", "secret", "token", "key"}

// scrub returns the word with any secret replaced, and whether it changed.
func scrub(w string) (string, bool) {
	if name, value, ok := strings.Cut(w, "="); ok && value != "" {
		lower := strings.ToLower(name)
		for _, s := range sensitive {
			if strings.Contains(lower, s) {
				return name + "=" + redacted, true
			}
		}
	}
	if strings.HasPrefix(w, "sk_live_") && len(w) > 8 {
		return redacted, true
	}
	return w, false
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	count := 0
	for range n {
		sc.Scan()
		words := strings.Fields(sc.Text())
		prev := "" // the previous word as it was in the input
		for i, w := range words {
			after := prev == "Bearer"
			prev = w
			if after {
				words[i] = redacted
				count++
				continue
			}
			if s, changed := scrub(w); changed {
				words[i] = s
				count++
			}
		}
		fmt.Fprintln(out, strings.Join(words, " "))
	}
	fmt.Fprintln(out, "redacted", count)
}
