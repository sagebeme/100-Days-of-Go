package owasptop10

import (
	"io"
)

// Ticket is what the booking confirmation page shows.
type Ticket struct {
	Passenger  string
	Route      string
	Seat       int
	ProfileURL string
}

// RenderTicket writes the confirmation page for t. Every field is user input. The page is:
//
//	<h1>Tiketi ya {{.Passenger}}</h1>
//	<p>Route {{.Route}}, seat {{.Seat}}</p>
//	<a href="{{.ProfileURL}}">Profile</a>
func RenderTicket(w io.Writer, t Ticket) error {
	// Your code here
	return nil
}
