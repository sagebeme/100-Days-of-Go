package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	var k int64
	fmt.Fscan(in, &n, &k)

	// A run of days i+1..j adds up to K exactly when prefix[j] - prefix[i] == K. So at
	// each day, count the earlier prefix sums equal to prefix - K. One pass, O(N).
	seen := map[int64]int64{0: 1}
	var prefix, count int64
	for range n {
		var x int64
		fmt.Fscan(in, &x)
		prefix += x
		count += seen[prefix-k]
		seen[prefix]++
	}
	fmt.Println(count)
}
