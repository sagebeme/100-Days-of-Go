package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	fmt.Fscan(in, &n)
	total := 0
	for i := 0; i < n; i++ {
		var price int
		fmt.Fscan(in, &price)
		total += price
	}
	fmt.Println(total)
}
