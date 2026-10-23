// Package packageitup prices a shopping basket for a mama mboga stall. It uses the
// ksh package, in the ksh folder next to this file, to format the money.
package packageitup

// Total returns the cost in shillings of basket, which maps each item to how many the
// customer wants, using prices, which maps each item to its price. It returns an error
// if an item has no price or a quantity is below 1.
//
//	Total(map[string]int{"unga": 2}, map[string]int{"unga": 230}) → 460, nil
func Total(basket, prices map[string]int) (int, error) {
	// Your code here. Follow the steps in the README.
	return 0, nil
}

// Summary describes basket in one line: how many things it holds and what they cost,
// formatted with ksh.Format. It says "item" for exactly one thing and "items"
// otherwise. It returns the same errors as Total.
//
//	Summary(map[string]int{"unga": 2, "sukari": 1}, prices) → "3 items, Ksh645", nil
func Summary(basket, prices map[string]int) (string, error) {
	// Your code here. Import the ksh package and call ksh.Format.
	return "", nil
}
