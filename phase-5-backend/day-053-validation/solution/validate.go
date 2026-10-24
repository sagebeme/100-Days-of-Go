// Package validation checks every booking before it touches the store: the body
// is limited in size, each field is checked, and the client gets every problem at
// once, field by field.
package validation

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

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
	parts := make([]string, len(e))
	for i, fe := range e {
		parts[i] = fe.Field + " " + fe.Message
	}
	return strings.Join(parts, "; ")
}

// NormalizePhone returns a Kenyan mobile number as 2547XXXXXXXX or 2541XXXXXXXX,
// the way Day 9's Clean did. It accepts 0712 345 678, 0712-345-678, 712345678,
// 254712345678 and +254712345678. Spaces and hyphens are ignored, and a plus sign
// is allowed only at the very start. Anything else is an error.
func NormalizePhone(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty phone number")
	}
	plus := false
	var digits strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits.WriteByte(c)
		case c == ' ' || c == '-':
		case c == '+' && i == 0:
			plus = true
		default:
			return "", errors.New("phone number has a character that isn't a digit")
		}
	}
	d := digits.String()
	var local string
	switch {
	case len(d) == 12 && strings.HasPrefix(d, "254"):
		local = d[3:]
	case plus:
		return "", errors.New("a number starting with + must start with +254")
	case len(d) == 10 && d[0] == '0':
		local = d[1:]
	case len(d) == 9:
		local = d
	default:
		return "", errors.New("wrong number of digits for a Kenyan mobile number")
	}
	if local[0] != '7' && local[0] != '1' {
		return "", errors.New("a Kenyan mobile number starts 07 or 01")
	}
	return "254" + local, nil
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
	var errs Errors
	add := func(field, msg string) { errs = append(errs, FieldError{Field: field, Message: msg}) }

	req.Passenger = strings.TrimSpace(req.Passenger)
	switch n := utf8.RuneCountInString(req.Passenger); {
	case n == 0:
		add("passenger", "is required")
	case n < 2 || n > 60:
		add("passenger", "must be 2 to 60 characters")
	case !nameOK(req.Passenger):
		add("passenger", "may only contain letters, spaces, hyphens, apostrophes and full stops")
	}

	if strings.TrimSpace(req.Phone) == "" {
		add("phone", "is required")
	} else if phone, err := NormalizePhone(req.Phone); err != nil {
		add("phone", "must be a Kenyan mobile number like 0712 345 678")
	} else {
		req.Phone = phone
	}

	if req.Seat < 1 || req.Seat > trip.Seats {
		add("seat", "must be between 1 and "+strconv.Itoa(trip.Seats))
	}

	if !trip.Departs.After(now) {
		add("trip", "has already left")
	}
	return req, errs
}

// nameOK allows letters from any alphabet, so Wanjirũ and Nyokabi are both fine, plus
// the punctuation real names use. It rules out digits, symbols, control characters
// and everything else.
func nameOK(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' && r != '.' {
			return false
		}
	}
	return true
}
