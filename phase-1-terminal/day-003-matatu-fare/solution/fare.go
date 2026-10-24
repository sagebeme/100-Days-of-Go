// Package matatufare works out matatu fares and route numbers.
package matatufare

// Fare returns the fare in shillings for a trip of distanceKm kilometres that starts
// at hour (0 to 23). The README has the full table of rules.
//
//	Fare(3, 10, false) → 50
//	Fare(12, 7, false) → 110
//	Fare(40, 18, true) → 250
func Fare(distanceKm, hour int, raining bool) int {
	if distanceKm <= 0 {
		return 0
	}

	var fare int
	switch {
	case distanceKm <= 5:
		fare = 50
	case distanceKm <= 15:
		fare = 80
	case distanceKm <= 30:
		fare = 120
	default:
		fare = 200
	}

	morningRush := hour >= 6 && hour <= 8
	eveningRush := hour >= 17 && hour <= 19
	if morningRush || eveningRush {
		fare += 30
	}
	if raining {
		fare += 20
	}
	return fare
}

// Route returns the route number of the matatu that goes to destination,
// or "unknown" if it isn't one we know.
//
//	Route("Rongai")  → "125"
//	Route("Mombasa") → "unknown"
func Route(destination string) string {
	switch destination {
	case "Kawangware":
		return "46K"
	case "Ngong":
		return "111"
	case "Rongai", "Ongata Rongai":
		return "125"
	case "Githurai":
		return "45"
	case "Buruburu":
		return "58"
	default:
		return "unknown"
	}
}
