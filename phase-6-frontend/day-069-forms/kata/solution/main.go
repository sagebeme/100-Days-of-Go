package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var strip = strings.NewReplacer(" ", "", "-", "")

// normalise returns 254XXXXXXXXX, or "" if s isn't a Kenyan mobile number.
func normalise(s string) string {
	s = strip.Replace(s)
	var rest string
	switch {
	case len(s) == 10 && s[0] == '0':
		rest = s[1:]
	case len(s) == 12 && strings.HasPrefix(s, "254"):
		rest = s[3:]
	case len(s) == 13 && strings.HasPrefix(s, "+254"):
		rest = s[4:]
	default:
		return ""
	}
	if rest[0] != '7' && rest[0] != '1' {
		return ""
	}
	for _, c := range rest {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return "254" + rest
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	seen := map[string]bool{}
	for i := 0; i < n && sc.Scan(); i++ {
		p := normalise(sc.Text())
		if p == "" {
			fmt.Fprintln(out, "INVALID")
			continue
		}
		seen[p] = true
		fmt.Fprintln(out, p)
	}
	fmt.Fprintf(out, "unique %d\n", len(seen))
}
