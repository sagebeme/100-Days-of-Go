package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var k int
	fmt.Fscan(in, &k)
	want := 0
	for range k {
		var c int
		fmt.Fscan(in, &c)
		want += c
	}

	reg := make([]int, k+1)
	step := make([]int, k+1)
	x := 0
	var steps int
	fmt.Fscan(in, &steps)
	for range steps {
		var g int
		fmt.Fscan(in, &g)
		switch step[g] {
		case 0:
			reg[g] = x
		case 1:
			reg[g]++
		case 2:
			x = reg[g]
		}
		step[g] = (step[g] + 1) % 3
	}
	fmt.Println(x)
	fmt.Println(want - x)
}
