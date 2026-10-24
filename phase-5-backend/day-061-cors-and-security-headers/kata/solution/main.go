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

	var n, w, k int
	fmt.Fscan(in, &n, &w, &k)
	// log[ip] holds the times of that IP's allowed requests still inside the window,
	// oldest first. Times never decrease, so old ones are always at the front.
	log := map[string][]int{}
	for i := 0; i < n; i++ {
		var t int
		var ip string
		fmt.Fscan(in, &t, &ip)
		times := log[ip]
		for len(times) > 0 && times[0] <= t-w {
			times = times[1:]
		}
		if len(times) < k {
			times = append(times, t)
			fmt.Fprintln(out, "ALLOW")
		} else {
			fmt.Fprintln(out, "DENY")
		}
		log[ip] = times
	}
}
