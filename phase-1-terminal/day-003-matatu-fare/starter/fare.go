// Package matatufare works out matatu fares and route numbers.
package matatufare

// Fare returns the fare in shillings for a trip of distanceKm kilometres that starts
// at hour (0 to 23). The README has the full table of rules.
//
//	Fare(3, 10, false) → 50
//	Fare(12, 7, false) → 110
//	Fare(40, 18, true) → 250
func Fare(distanceKm, hour int, raining bool) int {
	// Your code here
	return 0
}

// Route returns the route number of the matatu that goes to destination,
// or "unknown" if it isn't one we know.
//
//	Route("Rongai")  → "125"
//	Route("Mombasa") → "unknown"
func Route(destination string) string {
	// Your code here
	return ""
}
