// Package owasptop10 fixes one OWASP Top 10 risk per file in the Twende booking API:
// injection, broken access control, SSRF, XSS and secrets in logs.
package owasptop10

import (
	"errors"
	"strconv"
	"strings"
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
	col, ok := sortColumns[f.Sort]
	if !ok {
		return "", nil, ErrBadSort
	}

	var where []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, cond+" $"+strconv.Itoa(len(args)))
	}
	if f.RouteID != 0 {
		add("t.route_id =", f.RouteID)
	}
	if !f.From.IsZero() {
		add("t.departs >=", f.From)
	}
	if f.Name != "" {
		add("r.name ILIKE", "%"+f.Name+"%")
	}

	var b strings.Builder
	b.WriteString("SELECT t.id, t.route_id, t.departs, t.seats FROM trips t JOIN routes r ON r.id = t.route_id")
	if len(where) > 0 {
		b.WriteString(" WHERE " + strings.Join(where, " AND "))
	}
	b.WriteString(" ORDER BY " + col)
	if f.Desc {
		b.WriteString(" DESC")
	} else {
		b.WriteString(" ASC")
	}

	limit := f.Limit
	switch {
	case limit <= 0:
		limit = 20
	case limit > 100:
		limit = 100
	}
	args = append(args, limit)
	b.WriteString(" LIMIT $" + strconv.Itoa(len(args)))
	return b.String(), args, nil
}
