// Package validation checks every booking before it touches the store: the body
// is limited in size, each field is checked, and the client gets every problem at
// once, field by field.
package validation

import "time"

// FieldError is one problem with one field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Errors is every problem with a request, in the order the fields were checked.
type Errors []FieldError

// Error joins the problems as "field message; field message".
//
//	Errors{{"seat", "must be between 1 and 14"}, {"phone", "is required"}}.Error()
//	  →  "seat must be between 1 and 14; phone is required"
func (e Errors) Error() string {
	// Your code here.
	return ""
}

// NormalizePhone returns a Kenyan mobile number as 2547XXXXXXXX or 2541XXXXXXXX,
// the way Day 9's Clean did. It accepts 0712 345 678, 0712-345-678, 712345678,
// 254712345678 and +254712345678. Spaces and hyphens are ignored, and a plus sign
// is allowed only at the very start. Anything else is an error.
func NormalizePhone(raw string) (string, error) {
	// Your code here. Bring over your Day 9 Clean.
	return "", nil
}

// BookingRequest is the body of POST /trips/{id}/bookings.
type BookingRequest struct {
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// ValidateBooking checks req for trip at time now. It returns req cleaned up (the
// passenger's name trimmed, the phone normalised) and every problem, checked in
// this order:
//
//	passenger  "is required" when blank
//	           "must be 2 to 60 characters" (counting runes, after trimming)
//	           "may only contain letters, spaces, hyphens, apostrophes and full stops"
//	phone      "is required" when blank
//	           "must be a Kenyan mobile number like 0712 345 678"
//	seat       "must be between 1 and N", N being trip.Seats
//	trip       "has already left" unless trip.Departs is after now
//
// Each field gets at most one message. The returned Errors is nil when all is well.
func ValidateBooking(req BookingRequest, trip Trip, now time.Time) (BookingRequest, Errors) {
	// Your code here.
	return BookingRequest{}, nil
}
