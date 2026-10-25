package main

import (
	"bufio"
	"fmt"
	"os"
)

const mod = 1_000_000_007

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	var s string
	fmt.Fscan(in, &n, &s)

	// ways[i] is the number of ways to decode s[i:]. Fill it from the end.
	// Only ways[i+1] and ways[i+2] are needed, so two variables do.
	next1, next2 := 1, 0 // ways[n] = 1 (the empty rest), ways[n+1] unused
	for i := n - 1; i >= 0; i-- {
		cur := 0
		if s[i] != '0' {
			cur = next1 // s[i] alone is a letter, 1 to 9
			if i+1 < n && (s[i] == '1' || s[i] == '2' && s[i+1] <= '6') {
				cur = (cur + next2) % mod // s[i:i+2] is a letter, 10 to 26
			}
		}
		next1, next2 = cur, next1
	}
	fmt.Println(next1)
}
