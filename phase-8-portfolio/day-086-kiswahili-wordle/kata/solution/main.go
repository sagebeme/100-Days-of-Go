package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
)

func score(guess, answer string) string {
	marks := []byte(".....")
	var left [26]int
	for i := range 5 {
		if guess[i] == answer[i] {
			marks[i] = 'G'
		} else {
			left[answer[i]-'a']++
		}
	}
	for i := range 5 {
		if marks[i] == 'G' {
			continue
		}
		if c := guess[i] - 'a'; left[c] > 0 {
			marks[i] = 'Y'
			left[c]--
		}
	}
	return string(marks)
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	words := make([]string, n)
	for i := range words {
		fmt.Fscan(in, &words[i])
	}
	var m int
	fmt.Fscan(in, &m)
	guesses := make([]string, m)
	patterns := make([]string, m)
	for i := range m {
		fmt.Fscan(in, &guesses[i], &patterns[i])
	}

	var possible []string
	for _, w := range words {
		ok := true
		for i := range m {
			if score(guesses[i], w) != patterns[i] {
				ok = false
				break
			}
		}
		if ok {
			possible = append(possible, w)
		}
	}
	slices.Sort(possible)
	fmt.Fprintln(out, len(possible))
	for _, w := range possible[:min(10, len(possible))] {
		fmt.Fprintln(out, w)
	}
}
