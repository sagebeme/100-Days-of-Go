package main

import (
	"bufio"
	"os"
	"strconv"
)

// sortCount merge-sorts a (using buf as scratch space) and returns the number
// of inversions: pairs i < j with a[i] > a[j].
func sortCount(a, buf []int) int {
	if len(a) <= 1 {
		return 0
	}
	mid := len(a) / 2
	count := sortCount(a[:mid], buf[:mid]) + sortCount(a[mid:], buf[mid:])
	copy(buf, a)
	left, right := buf[:mid], buf[mid:len(a)]
	i, j, k := 0, 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			a[k] = left[i]
			i++
		} else {
			// right[j] jumps ahead of everything still waiting in left.
			a[k] = right[j]
			j++
			count += len(left) - i
		}
		k++
	}
	k += copy(a[k:], left[i:])
	copy(a[k:], right[j:])
	return count
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
	n := next()
	a := make([]int, n)
	for i := range a {
		a[i] = next()
	}
	os.Stdout.WriteString(strconv.Itoa(sortCount(a, make([]int, n))) + "\n")
}
