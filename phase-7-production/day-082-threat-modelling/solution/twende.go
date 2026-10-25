package threatmodelling

// TwendeModel is the matatu booking system from Days 59 to 81 as a data-flow diagram.
// It has one deliberate weakness for you to find: look at the flow from M-Pesa.
func TwendeModel() Model {
	return Model{
		Components: []Component{
			{Name: "Passenger app", Kind: External, Zone: "internet"},
			{Name: "M-Pesa", Kind: External, Zone: "partner"},
			{Name: "Booking API", Kind: Process, Zone: "dmz"},
			{Name: "Seat service", Kind: Process, Zone: "internal"},
			{Name: "Postgres", Kind: DataStore, Zone: "internal"},
			{Name: "Redis", Kind: DataStore, Zone: "internal"},
		},
		Flows: []Flow{
			{From: "Passenger app", To: "Booking API", Data: "login, trip searches, bookings", Encrypted: true, Authenticated: true},
			{From: "Booking API", To: "M-Pesa", Data: "STK push payment request", Encrypted: true, Authenticated: true},
			{From: "M-Pesa", To: "Booking API", Data: "payment confirmation callback", Encrypted: true, Authenticated: false},
			{From: "Booking API", To: "Seat service", Data: "BookSeat over gRPC", Encrypted: false, Authenticated: false},
			{From: "Booking API", To: "Redis", Data: "cached trips", Encrypted: false, Authenticated: true},
			{From: "Seat service", To: "Postgres", Data: "seats and bookings", Encrypted: true, Authenticated: true},
		},
	}
}
