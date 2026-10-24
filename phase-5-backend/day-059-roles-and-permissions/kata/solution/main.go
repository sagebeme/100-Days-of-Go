package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	grants := map[string]map[string]bool{} // role → actions
	parents := map[string][]string{}       // role → roles it inherits from
	users := map[string]string{}           // user → role

	var n int
	fmt.Fscan(in, &n)
	for i := 0; i < n; i++ {
		var kind, a, b string
		fmt.Fscan(in, &kind, &a, &b)
		switch kind {
		case "GRANT":
			if grants[a] == nil {
				grants[a] = map[string]bool{}
			}
			grants[a][b] = true
		case "INHERIT":
			parents[a] = append(parents[a], b)
		case "USER":
			users[a] = b
		}
	}

	// reach[role] is every role whose grants apply to role, itself included.
	// A seen set makes cycles harmless.
	reach := map[string][]string{}
	reachable := func(role string) []string {
		if r, ok := reach[role]; ok {
			return r
		}
		seen := map[string]bool{role: true}
		list := []string{role}
		for i := 0; i < len(list); i++ {
			for _, p := range parents[list[i]] {
				if !seen[p] {
					seen[p] = true
					list = append(list, p)
				}
			}
		}
		reach[role] = list
		return list
	}

	var q int
	fmt.Fscan(in, &q)
	for i := 0; i < q; i++ {
		var user, action string
		fmt.Fscan(in, &user, &action)
		verdict := "DENY"
		if role, ok := users[user]; ok {
			for _, r := range reachable(role) {
				if grants[r][action] {
					verdict = "ALLOW"
					break
				}
			}
		}
		fmt.Fprintln(out, verdict)
	}
}
