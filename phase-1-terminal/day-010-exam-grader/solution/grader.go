// Package examgrader works out class averages, grades and the order of merit for a set of exam scores.
package examgrader

import (
	"fmt"
	"sort"
)

// Mean returns the average of scores, or 0 when there are none.
//
//	Mean([]int{70, 80, 90}) → 80
func Mean(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0
	for _, s := range scores {
		total += s
	}
	return float64(total) / float64(len(scores))
}

// Median returns the middle score once the scores are sorted. With an even number of
// scores it returns the average of the two in the middle. It returns 0 when there are
// none, and it never changes the slice it is given.
//
//	Median([]int{90, 40, 70}) → 70
func Median(scores []int) float64 {
	n := len(scores)
	if n == 0 {
		return 0
	}
	sorted := make([]int, n)
	copy(sorted, scores)
	sort.Ints(sorted)
	if n%2 == 1 {
		return float64(sorted[n/2])
	}
	return float64(sorted[n/2-1]+sorted[n/2]) / 2
}

// Grade turns a score from 0 to 100 into a letter: A for 80 and above, B for 65 to 79,
// C for 50 to 64, D for 40 to 49 and E below 40. A score outside 0 to 100 is an error.
//
//	Grade(72) → "B", nil
func Grade(score int) (string, error) {
	switch {
	case score < 0 || score > 100:
		return "", fmt.Errorf("score %d is outside 0 to 100", score)
	case score >= 80:
		return "A", nil
	case score >= 65:
		return "B", nil
	case score >= 50:
		return "C", nil
	case score >= 40:
		return "D", nil
	default:
		return "E", nil
	}
}

// Rank returns the names in scores from the highest score to the lowest. Students with
// the same score are listed in alphabetical order.
//
//	Rank(map[string]int{"Amina": 70, "Baraka": 85, "Chebet": 70}) → ["Baraka", "Amina", "Chebet"]
func Rank(scores map[string]int) []string {
	names := make([]string, 0, len(scores))
	for name := range scores {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		a, b := names[i], names[j]
		if scores[a] != scores[b] {
			return scores[a] > scores[b]
		}
		return a < b
	})
	return names
}
