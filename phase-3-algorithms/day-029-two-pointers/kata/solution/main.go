package main

import (
	"bufio"
	"fmt"
	"os"
)

// isPal reports whether s[i..j] is a palindrome.
func isPal(s string, i, j int) bool {
	for i < j {
		if s[i] != s[j] {
			return false
		}
		i++
		j--
	}
	return true
}

// almost reports whether s is a palindrome after deleting at most one character.
func almost(s string) bool {
	i, j := 0, len(s)-1
	for i < j && s[i] == s[j] {
		i++
		j--
	}
	if i >= j {
		return true
	}
	// First mismatch: one of these two must go.
	return isPal(s, i+1, j) || isPal(s, i, j-1)
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q int
	fmt.Fscan(in, &q)
	for range q {
		var s string
		fmt.Fscan(in, &s)
		if almost(s) {
			fmt.Fprintln(out, "YES")
		} else {
			fmt.Fprintln(out, "NO")
		}
	}
}
