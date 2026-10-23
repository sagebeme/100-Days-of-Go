// Package examgrader works out class averages, grades and the order of merit for a set of exam scores.
package examgrader

// Mean returns the average of scores, or 0 when there are none.
//
//	Mean([]int{70, 80, 90}) → 80
func Mean(scores []int) float64 {
	// Your code here. Follow the steps in the README.
	return 0
}

// Median returns the middle score once the scores are sorted. With an even number of
// scores it returns the average of the two in the middle. It returns 0 when there are
// none, and it never changes the slice it is given.
//
//	Median([]int{90, 40, 70}) → 70
func Median(scores []int) float64 {
	// Your code here.
	return 0
}

// Grade turns a score from 0 to 100 into a letter: A for 80 and above, B for 65 to 79,
// C for 50 to 64, D for 40 to 49 and E below 40. A score outside 0 to 100 is an error.
//
//	Grade(72) → "B", nil
func Grade(score int) (string, error) {
	// Your code here.
	return "", nil
}

// Rank returns the names in scores from the highest score to the lowest. Students with
// the same score are listed in alphabetical order.
//
//	Rank(map[string]int{"Amina": 70, "Baraka": 85, "Chebet": 70}) → ["Baraka", "Amina", "Chebet"]
func Rank(scores map[string]int) []string {
	// Your code here.
	return nil
}
