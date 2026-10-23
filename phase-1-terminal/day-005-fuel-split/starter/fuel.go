// Package fuelsplit works out what a road trip's fuel costs and how to share it.
package fuelsplit

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
	// Your code here
	return 0, nil
}

// Split shares cost equally among people. It returns each person's share and the
// shillings left over, which the car's owner pays.
//
// It returns an error if people is less than 1 or cost is negative.
//
//	Split(738, 4) → 184, 2, nil
//	Split(500, 0) → 0, 0, error
func Split(cost, people int) (share, leftover int, err error) {
	// Your code here
	return 0, 0, nil
}
