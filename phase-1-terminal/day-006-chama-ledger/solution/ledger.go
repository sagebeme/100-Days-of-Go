// Package chamaledger keeps the books for a chama's monthly contributions.
package chamaledger

// Total returns the sum of all contributions, or 0 if there are none.
//
//	Total([]int{1000, 1000, 500}) → 2500
func Total(contributions []int) int {
	total := 0
	for _, c := range contributions {
		total += c
	}
	return total
}

// RunningBalance returns the balance after each contribution, in order.
// It doesn't change contributions.
//
//	RunningBalance([]int{500, 1000, 250}) → [500 1500 1750]
func RunningBalance(contributions []int) []int {
	var balances []int
	sum := 0
	for _, c := range contributions {
		sum += c
		balances = append(balances, sum)
	}
	return balances
}

// Defaulters returns the names of members who paid less than due, in the order
// they appear. names[i] paid paid[i], and both slices are the same length.
//
//	Defaulters([]string{"Njeri", "Atieno", "Wafula"}, []int{1000, 600, 1000}, 1000) → [Atieno]
func Defaulters(names []string, paid []int, due int) []string {
	var owing []string
	for i, name := range names {
		if paid[i] < due {
			owing = append(owing, name)
		}
	}
	return owing
}
