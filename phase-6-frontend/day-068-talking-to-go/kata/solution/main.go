package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	fmt.Fscan(in, &n)
	sent := make([]int, n)
	arrives := make([]int, n)
	for i := range n {
		var d int
		fmt.Fscan(in, &sent[i], &d)
		arrives[i] = sent[i] + d
	}

	var shown []string
	last := 0 // the request whose answer lands last without abort
	for i := range n {
		if i == n-1 || arrives[i] < sent[i+1] {
			shown = append(shown, strconv.Itoa(i+1))
		}
		if arrives[i] >= arrives[last] {
			last = i
		}
	}
	fmt.Printf("shown %s\nwithout abort %d\n", strings.Join(shown, " "), last+1)
}
