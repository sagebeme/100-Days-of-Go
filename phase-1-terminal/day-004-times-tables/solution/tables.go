// Package timestables prints times tables and answers Mwalimu's bonus questions.
package timestables

import "fmt"

// Table returns the n times table from n x 1 up to n x upTo, one line per row,
// each line ending in a newline. It returns "" when upTo is less than 1.
//
//	Table(3, 3) → "3 x 1 = 3\n3 x 2 = 6\n3 x 3 = 9\n"
func Table(n, upTo int) string {
	text := ""
	for i := 1; i <= upTo; i++ {
		text += fmt.Sprintf("%d x %d = %d\n", n, i, n*i)
	}
	return text
}

// SumTable returns n x 1 + n x 2 + ... + n x upTo, or 0 when upTo is less than 1.
//
//	SumTable(7, 12) → 546
func SumTable(n, upTo int) int {
	total := 0
	for i := 1; i <= upTo; i++ {
		total += n * i
	}
	return total
}

// CountDivisible returns how many whole numbers from from to to (both included)
// divide exactly by d. d is always more than 0. It returns 0 when from is bigger than to.
//
//	CountDivisible(1, 100, 3) → 33
func CountDivisible(from, to, d int) int {
	count := 0
	for x := from; x <= to; x++ {
		if x%d == 0 {
			count++
		}
	}
	return count
}
