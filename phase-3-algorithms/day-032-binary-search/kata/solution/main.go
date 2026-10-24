package main

import (
	"bufio"
	"os"
	"strconv"
)

// daysNeeded is how many days it takes to deliver every parcel, in order, with
// a boda that carries at most capacity kg a day.
func daysNeeded(weights []int, capacity int) int {
	days, load := 1, 0
	for _, w := range weights {
		if load+w > capacity {
			days++
			load = 0
		}
		load += w
	}
	return days
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	sc.Split(bufio.ScanWords)
	next := func() int {
		sc.Scan()
		x, _ := strconv.Atoi(sc.Text())
		return x
	}
	n, d := next(), next()
	weights := make([]int, n)
	lo, hi := 0, 0 // the answer is between the heaviest parcel and the total
	for i := range weights {
		weights[i] = next()
		lo = max(lo, weights[i])
		hi += weights[i]
	}
	// More capacity never needs more days, so binary search for the smallest that fits.
	for lo < hi {
		mid := lo + (hi-lo)/2
		if daysNeeded(weights, mid) <= d {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	os.Stdout.WriteString(strconv.Itoa(lo) + "\n")
}
