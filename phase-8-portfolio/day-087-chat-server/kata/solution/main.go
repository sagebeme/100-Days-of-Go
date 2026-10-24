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

	var n int
	fmt.Fscan(in, &n)
	rooms := map[string]map[string]struct{}{}
	total := 0
	for range n {
		var op, user, room string
		fmt.Fscan(in, &op, &user, &room)
		if rooms[room] == nil {
			rooms[room] = map[string]struct{}{}
		}
		members := rooms[room]
		switch op {
		case "JOIN":
			members[user] = struct{}{}
		case "LEAVE":
			delete(members, user)
		case "SAY":
			if _, ok := members[user]; !ok {
				fmt.Fprintln(out, "denied")
				continue
			}
			fmt.Fprintln(out, len(members)-1)
			total += len(members) - 1
		}
	}
	fmt.Fprintln(out, total)
}
