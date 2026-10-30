package routes

// NairobiLinks is a small Nairobi network. The stops, route numbers and where each
// route boards in the CBD were checked on https://nishukishe.com/directions in
// October 2026. The links between stops and every travel time are SIMPLIFIED for
// this exercise: real matatus stop in many more places, and the times depend on
// traffic. Routes change, so confirm with the conductor.
var NairobiLinks = []Link{
	// Walking between CBD stages (rough times).
	{"Kencom", "Ambassadeur", Walk, 3},
	{"Ambassadeur", "Ronald Ngala Street", Walk, 6},
	{"Kencom", "Railways", Walk, 8},
	{"Ambassadeur", "Railways", Walk, 10},

	// Matatu routes, from where they board in the CBD to the end of the line.
	{"Kencom", "Kawangware", "46K", 45},                 // Citi Hoppa, boarding at Kencom
	{"Ambassadeur", "Kawangware", "46K", 45},            // Citi Hoppa, boarding at Ambassadeur
	{"Railways", "Ngong", "111", 70},                    // MCA Sacco
	{"Railways", "Ongata Rongai", "125", 60},            // City Shuttle
	{"Ronald Ngala Street", "Githurai", "45", 50},       // Zamzam
	{"Ambassadeur", "Buruburu", "58", 35},               // Buruburu 58 Travellers Sacco
	{"Railways", "Kibera", "Highrise Kibera Sacco", 30}, // route number not shown
}

// Nairobi returns the network built from NairobiLinks.
func Nairobi() *Network {
	n, err := NewNetwork(NairobiLinks)
	if err != nil {
		panic(err) // the data above is fixed, so this only fails if someone breaks it
	}
	return n
}
