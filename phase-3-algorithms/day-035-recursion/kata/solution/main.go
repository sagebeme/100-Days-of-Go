package main

import (
	"bufio"
	"fmt"
	"os"
)

const mod = 1_000_000_007

// power returns a^b mod m in O(log b): a^b is (a^(b/2))², times a once more if b is odd.
func power(a, b, m int) int {
	if b == 0 {
		return 1 % m
	}
	half := power(a, b/2, m)
	result := half * half % m // both are below 10^9+7, so the product fits in 64 bits
	if b%2 == 1 {
		result = result * (a % m) % m
	}
	return result
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var q int
	fmt.Fscan(in, &q)
	for range q {
		var a, b int
		fmt.Fscan(in, &a, &b)
		fmt.Fprintln(out, power(a, b, mod))
	}
}
