package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// roleSet holds up to 128 roles, one bit each. The kata promises at most 100.
type roleSet [2]uint64

func (s *roleSet) add(id int)          { s[id/64] |= 1 << (id % 64) }
func (s roleSet) has(id int) bool      { return s[id/64]&(1<<(id%64)) != 0 }
func (s roleSet) meets(t roleSet) bool { return s[0]&t[0] != 0 || s[1]&t[1] != 0 }

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	roleID := map[string]int{}
	var parents [][]int // role → roles it inherits from
	id := func(role string) int {
		if n, ok := roleID[role]; ok {
			return n
		}
		roleID[role] = len(parents)
		parents = append(parents, nil)
		return roleID[role]
	}
	grantedTo := map[string]roleSet{} // action → roles granted it directly
	users := map[string]int{}         // user → role

	line := func() []string { sc.Scan(); return strings.Fields(sc.Text()) }
	n, _ := strconv.Atoi(line()[0])
	for i := 0; i < n; i++ {
		f := line()
		switch f[0] {
		case "GRANT":
			s := grantedTo[f[2]]
			s.add(id(f[1]))
			grantedTo[f[2]] = s
		case "INHERIT":
			a, b := id(f[1]), id(f[2])
			parents[a] = append(parents[a], b)
		case "USER":
			users[f[1]] = id(f[2])
		}
	}
	// reach[r] is every role whose grants apply to r, itself included: a BFS per
	// role, at most 100 roles, and a loop is harmless because a role is only added once.
	reach := make([]roleSet, len(parents))
	for r := range reach {
		reach[r].add(r)
		queue := []int{r}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, p := range parents[cur] {
				if !reach[r].has(p) {
					reach[r].add(p)
					queue = append(queue, p)
				}
			}
		}
	}

	// Each question is now one map lookup and one AND of two 128-bit sets.
	q, _ := strconv.Atoi(line()[0])
	for i := 0; i < q; i++ {
		f := line()
		verdict := "DENY"
		if role, ok := users[f[0]]; ok && reach[role].meets(grantedTo[f[1]]) {
			verdict = "ALLOW"
		}
		out.WriteString(verdict)
		out.WriteByte('\n')
	}
}
