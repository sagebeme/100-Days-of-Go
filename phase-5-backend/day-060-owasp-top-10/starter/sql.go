// Package owasptop10 fixes one OWASP Top 10 risk per file in the Twende booking API:
// injection, broken access control, SSRF, XSS and secrets in logs.
package owasptop10

import (
	"errors"
	"time"
)

// ErrBadSort is returned for a sort column that isn't on the allowlist.
var ErrBadSort = errors.New("sort must be departs or seats")

// TripFilter is what a passenger can search trips by. Zero values mean "any".
type TripFilter struct {
	RouteID int       // only trips on this route
	From    time.Time // only trips departing at or after this time
	Name    string    // route name contains this text, any case
	Sort    string    // "departs" (the default) or "seats"
	Desc    bool      // newest or biggest first
	Limit   int       // 1 to 100; 0 means 20, above 100 means 100
}

// sortColumns is the allowlist. Placeholders work for values, not for column names,
// so a name the user picks must come from a fixed list.
var sortColumns = map[string]string{
	"":        "t.departs",
	"departs": "t.departs",
	"seats":   "t.seats",
}

// BuildTripSearch returns a Postgres query and its arguments for f. Every value the
// user sent goes in args and appears in the SQL only as $1, $2 and so on.
func BuildTripSearch(f TripFilter) (string, []any, error) {
	// Your code here
	return "", nil, nil
}
