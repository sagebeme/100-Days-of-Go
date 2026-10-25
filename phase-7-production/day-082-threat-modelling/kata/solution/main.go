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

	var n, m, k int
	fmt.Fscan(in, &n, &m, &k)
	index := make(map[string]int, n)
	dist := make([]int, n)
	var queue []int
	for i := range n {
		var name, zone string
		fmt.Fscan(in, &name, &zone)
		index[name] = i
		dist[i] = -1
		if zone == "internet" {
			dist[i] = 0
			queue = append(queue, i)
		}
	}
	edges := make([][]int, n)
	for range m {
		var from, to string
		fmt.Fscan(in, &from, &to)
		edges[index[from]] = append(edges[index[from]], index[to])
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range edges[u] {
			if dist[v] == -1 {
				dist[v] = dist[u] + 1
				queue = append(queue, v)
			}
		}
	}
	for range k {
		var name string
		fmt.Fscan(in, &name)
		if d := dist[index[name]]; d >= 0 {
			fmt.Fprintln(out, name, d)
		} else {
			fmt.Fprintln(out, name, "SAFE")
		}
	}
}
