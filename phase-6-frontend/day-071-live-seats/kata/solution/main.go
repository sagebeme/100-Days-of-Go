package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var s, v, m int
	var snapshot string
	fmt.Fscan(in, &s, &v, &m, &snapshot)
	seats := []byte(snapshot)
	versions := make([]int, s)
	for i := range versions {
		versions[i] = v
	}
	ignored := 0
	for i := 0; i < m; i++ {
		var version, seat, taken int
		fmt.Fscan(in, &version, &seat, &taken)
		if version <= versions[seat-1] {
			ignored++
			continue
		}
		versions[seat-1] = version
		seats[seat-1] = '0' + byte(taken)
	}
	fmt.Printf("%s\nignored %d\n", seats, ignored)
}
