// Package fuelsplit works out what a road trip's fuel costs and how to share it.
package fuelsplit

import (
	"errors"
	"fmt"
)

// TripCost returns the fuel cost in shillings of driving distanceKm kilometres in a
// car that does kmPerLitre km on one litre, with fuel at pricePerLitre shillings.
// The cost is rounded up to the next whole shilling.
//
// It returns an error if kmPerLitre is 0 or less, or if distanceKm or pricePerLitre
// is negative.
//
//	TripCost(100, 15, 180) → 1200, nil
//	TripCost(50, 12, 177)  → 738, nil
//	TripCost(10, 0, 180)   → 0, error
func TripCost(distanceKm, kmPerLitre, pricePerLitre int) (int, error) {
	if kmPerLitre <= 0 {
		return 0, fmt.Errorf("km per litre must be more than 0, got %d", kmPerLitre)
	}
	if distanceKm < 0 {
		return 0, fmt.Errorf("distance can't be negative, got %d", distanceKm)
	}
	if pricePerLitre < 0 {
		return 0, fmt.Errorf("price per litre can't be negative, got %d", pricePerLitre)
	}
	cost := (distanceKm*pricePerLitre + kmPerLitre - 1) / kmPerLitre
	return cost, nil
}

// Split shares cost equally among people. It returns each person's share and the
// shillings left over, which the car's owner pays.
//
// It returns an error if people is less than 1 or cost is negative.
//
//	Split(738, 4) → 184, 2, nil
//	Split(500, 0) → 0, 0, error
func Split(cost, people int) (share, leftover int, err error) {
	if people < 1 {
		return 0, 0, errors.New("need at least 1 person to split the cost")
	}
	if cost < 0 {
		return 0, 0, errors.New("cost can't be negative")
	}
	return cost / people, cost % people, nil
}
