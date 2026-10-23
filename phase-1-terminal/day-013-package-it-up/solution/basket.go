// Package packageitup prices a shopping basket for a mama mboga stall. It uses the
// ksh package, in the ksh folder next to this file, to format the money.
package packageitup

import (
	"fmt"

	"github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-013-package-it-up/solution/ksh"
)

// Total returns the cost in shillings of basket, which maps each item to how many the
// customer wants, using prices, which maps each item to its price. It returns an error
// if an item has no price or a quantity is below 1.
//
//	Total(map[string]int{"unga": 2}, map[string]int{"unga": 230}) → 460, nil
func Total(basket, prices map[string]int) (int, error) {
	total := 0
	for item, qty := range basket {
		price, ok := prices[item]
		if !ok {
			return 0, fmt.Errorf("no price for %q", item)
		}
		if qty < 1 {
			return 0, fmt.Errorf("%s: quantity %d must be at least 1", item, qty)
		}
		total += qty * price
	}
	return total, nil
}

// Summary describes basket in one line: how many things it holds and what they cost,
// formatted with ksh.Format. It says "item" for exactly one thing and "items"
// otherwise. It returns the same errors as Total.
//
//	Summary(map[string]int{"unga": 2, "sukari": 1}, prices) → "3 items, Ksh645", nil
func Summary(basket, prices map[string]int) (string, error) {
	total, err := Total(basket, prices)
	if err != nil {
		return "", err
	}
	count := 0
	for _, qty := range basket {
		count += qty
	}
	word := "items"
	if count == 1 {
		word = "item"
	}
	return fmt.Sprintf("%d %s, %s", count, word, ksh.Format(total)), nil
}
