package owasptop10

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

const base = "SELECT t.id, t.route_id, t.departs, t.seats FROM trips t JOIN routes r ON r.id = t.route_id"

func TestBuildTripSearch(t *testing.T) {
	from := time.Date(2026, 3, 2, 6, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		f         TripFilter
		wantQuery string
		wantArgs  []any
	}{
		{"no filter", TripFilter{},
			base + " ORDER BY t.departs ASC LIMIT $1", []any{20}},
		{"route only", TripFilter{RouteID: 111},
			base + " WHERE t.route_id = $1 ORDER BY t.departs ASC LIMIT $2", []any{111, 20}},
		{"everything", TripFilter{RouteID: 125, From: from, Name: "rongai", Sort: "seats", Desc: true, Limit: 5},
			base + " WHERE t.route_id = $1 AND t.departs >= $2 AND r.name ILIKE $3 ORDER BY t.seats DESC LIMIT $4",
			[]any{125, from, "%rongai%", 5}},
		{"name and time", TripFilter{From: from, Name: "Ngong"},
			base + " WHERE t.departs >= $1 AND r.name ILIKE $2 ORDER BY t.departs ASC LIMIT $3",
			[]any{from, "%Ngong%", 20}},
		{"limit capped at 100", TripFilter{Limit: 5000},
			base + " ORDER BY t.departs ASC LIMIT $1", []any{100}},
		{"negative limit means the default", TripFilter{Limit: -1, Sort: "departs"},
			base + " ORDER BY t.departs ASC LIMIT $1", []any{20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, args, err := BuildTripSearch(tt.f)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if q != tt.wantQuery {
				t.Errorf("query\n got: %s\nwant: %s", q, tt.wantQuery)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args\n got: %#v\nwant: %#v", args, tt.wantArgs)
			}
		})
	}
}

func TestBuildTripSearchInjection(t *testing.T) {
	attacks := []string{
		"' OR '1'='1",
		"x'; DROP TABLE bookings; --",
		"rongai\") UNION SELECT phone, password_hash, 1, 1 FROM users --",
	}
	for _, a := range attacks {
		q, args, err := BuildTripSearch(TripFilter{Name: a})
		if err != nil {
			t.Fatalf("BuildTripSearch(Name: %q) error: %v", a, err)
		}
		if strings.Contains(q, a) || strings.Contains(q, "DROP") || strings.Contains(q, "UNION") || strings.Contains(q, "'") {
			t.Errorf("user input reached the SQL text:\n%s", q)
		}
		if len(args) == 0 || args[0] != "%"+a+"%" {
			t.Errorf("args = %#v, want the input as the first argument, wrapped in %%", args)
		}
	}
}

func TestBuildTripSearchSortAllowlist(t *testing.T) {
	for _, s := range []string{"departs; DROP TABLE trips", "id", "t.seats", "SEATS", "seats DESC"} {
		q, args, err := BuildTripSearch(TripFilter{Sort: s})
		if !errors.Is(err, ErrBadSort) || q != "" || args != nil {
			t.Errorf("BuildTripSearch(Sort: %q)\n got: %q, %v, %v\nwant: \"\", nil, ErrBadSort", s, q, args, err)
		}
	}
}
