package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type pattern struct {
	method string
	segs   []string
}

func segments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func matches(pat, path []string) bool {
	if len(pat) != len(path) {
		return false
	}
	for i, s := range pat {
		wild := strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
		if !wild && s != path[i] {
			return false
		}
	}
	return true
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	pats := make([]pattern, n)
	for i := range pats {
		var m, p string
		fmt.Fscan(in, &m, &p)
		pats[i] = pattern{m, segments(p)}
	}

	var q int
	fmt.Fscan(in, &q)
	for ; q > 0; q-- {
		var m, p string
		fmt.Fscan(in, &m, &p)
		path := segments(p)
		allowed := map[string]bool{}
		for _, pat := range pats {
			if matches(pat.segs, path) {
				allowed[pat.method] = true
				if pat.method == "GET" {
					allowed["HEAD"] = true
				}
			}
		}
		switch {
		case allowed[m]:
			fmt.Fprintln(out, "200")
		case len(allowed) > 0:
			names := make([]string, 0, len(allowed))
			for name := range allowed {
				names = append(names, name)
			}
			sort.Strings(names)
			fmt.Fprintln(out, "405 "+strings.Join(names, ", "))
		default:
			fmt.Fprintln(out, "404")
		}
	}
}
