package owasptop10

import (
	"html/template"
	"io"
)

// Ticket is what the booking confirmation page shows.
type Ticket struct {
	Passenger  string
	Route      string
	Seat       int
	ProfileURL string
}

// ticketPage is parsed by html/template, which escapes each value for where it
// appears: HTML text, an attribute, a URL.
var ticketPage = template.Must(template.New("ticket").Parse(
	`<h1>Tiketi ya {{.Passenger}}</h1>
<p>Route {{.Route}}, seat {{.Seat}}</p>
<a href="{{.ProfileURL}}">Profile</a>
`))

// RenderTicket writes the confirmation page for t. Every field is user input.
func RenderTicket(w io.Writer, t Ticket) error {
	return ticketPage.Execute(w, t)
}
