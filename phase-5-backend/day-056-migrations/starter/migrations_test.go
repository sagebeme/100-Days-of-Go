package migrations

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

// good is a folder of migrations, listed out of order on purpose.
var good = fstest.MapFS{
	"db/0002_create_trips.up.sql":      file("CREATE TABLE trips (id bigserial PRIMARY KEY);"),
	"db/0002_create_trips.down.sql":    file("DROP TABLE trips;"),
	"db/0001_create_routes.up.sql":     file("CREATE TABLE routes (id bigserial PRIMARY KEY);"),
	"db/0001_create_routes.down.sql":   file("DROP TABLE routes;"),
	"db/0010_add_fare.up.sql":          file("ALTER TABLE routes ADD COLUMN fare integer;"),
	"db/0010_add_fare.down.sql":        file("ALTER TABLE routes DROP COLUMN fare;"),
	"db/README.md":                     file("Migrations for the booking API."),
	"db/old/0099_ignored_dir.up.sql":   file("SELECT 1;"),
	"db/old/0099_ignored_dir.down.sql": file("SELECT 1;"),
}

func versions(ms []Migration) []int {
	out := []int{}
	for _, m := range ms {
		out = append(out, m.Version)
	}
	return out
}

func TestParse(t *testing.T) {
	ms, err := Parse(good, "db")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Migration{
		{Version: 1, Name: "create_routes", Up: "CREATE TABLE routes (id bigserial PRIMARY KEY);", Down: "DROP TABLE routes;"},
		{Version: 2, Name: "create_trips", Up: "CREATE TABLE trips (id bigserial PRIMARY KEY);", Down: "DROP TABLE trips;"},
		{Version: 10, Name: "add_fare", Up: "ALTER TABLE routes ADD COLUMN fare integer;", Down: "ALTER TABLE routes DROP COLUMN fare;"},
	}
	if !reflect.DeepEqual(ms, want) {
		t.Errorf("Parse\n got: %+v\nwant: %+v", ms, want)
	}

	empty := fstest.MapFS{"db/README.md": file("nothing yet")}
	ms, err = Parse(empty, "db")
	if err != nil || len(ms) != 0 {
		t.Errorf("Parse of a folder with no .sql files = %v, %v, want no migrations and no error", ms, err)
	}
	if _, err := Parse(good, "nope"); err == nil {
		t.Error("Parse of a folder that doesn't exist returned no error")
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		blame string // the error must mention this
	}{
		{"no direction", map[string]string{"0001_routes.sql": "x"}, "0001_routes.sql"},
		{"no version", map[string]string{"create_routes.up.sql": "x", "create_routes.down.sql": "x"}, "create_routes"},
		{"letters in the version", map[string]string{"v1_routes.up.sql": "x", "v1_routes.down.sql": "x"}, "v1_routes"},
		{"version zero", map[string]string{"0000_routes.up.sql": "x", "0000_routes.down.sql": "x"}, "0000_routes"},
		{"no name", map[string]string{"0001_.up.sql": "x", "0001_.down.sql": "x"}, "0001_"},
		{"capital letters", map[string]string{"0001_Routes.up.sql": "x", "0001_Routes.down.sql": "x"}, "0001_Routes"},
		{"a space", map[string]string{"0001_create routes.up.sql": "x", "0001_create routes.down.sql": "x"}, "0001_create routes"},
		{"missing down", map[string]string{"0001_routes.up.sql": "x", "0002_trips.up.sql": "x", "0002_trips.down.sql": "x"}, "1"},
		{"missing up", map[string]string{"0001_routes.down.sql": "x"}, "1"},
		{"two names, one version", map[string]string{"0001_routes.up.sql": "x", "0001_routes.down.sql": "x", "0001_trips.up.sql": "x", "0001_trips.down.sql": "x"}, "0001"},
		{"up and down disagree", map[string]string{"0001_routes.up.sql": "x", "0001_route.down.sql": "x"}, "0001"},
		{"same version written twice", map[string]string{"1_routes.up.sql": "x", "001_routes.up.sql": "x", "1_routes.down.sql": "x"}, "routes"},
		{"empty up", map[string]string{"0001_routes.up.sql": "  \n", "0001_routes.down.sql": "x"}, "0001_routes.up.sql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for name, body := range tt.files {
				fsys["m/"+name] = file(body)
			}
			ms, err := Parse(fsys, "m")
			if err == nil {
				t.Fatalf("Parse = %+v with no error, want an error", ms)
			}
			if !strings.Contains(err.Error(), tt.blame) {
				t.Errorf("error %q doesn't mention %q", err, tt.blame)
			}
			if ms != nil {
				t.Errorf("Parse returned %+v with its error, want nil", ms)
			}
		})
	}
}

var all = []Migration{
	{Version: 1, Name: "create_routes"},
	{Version: 2, Name: "create_trips"},
	{Version: 3, Name: "create_bookings"},
	{Version: 5, Name: "add_fare"},
}

func TestPending(t *testing.T) {
	tests := []struct {
		name    string
		applied []int
		want    []int
	}{
		{"fresh database", nil, []int{1, 2, 3, 5}},
		{"half way", []int{1, 2}, []int{3, 5}},
		{"applied out of order in the table", []int{2, 1}, []int{3, 5}},
		{"up to date", []int{1, 2, 3, 5}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Pending(all, tt.applied)
			if err != nil {
				t.Fatalf("Pending(%v) error = %v", tt.applied, err)
			}
			if got == nil {
				t.Fatalf("Pending(%v) = nil, want an empty slice when nothing is pending", tt.applied)
			}
			if g := versions(got); !reflect.DeepEqual(g, tt.want) {
				t.Errorf("Pending(%v) = %v, want %v", tt.applied, g, tt.want)
			}
		})
	}
}

func TestPendingRefuses(t *testing.T) {
	_, err := Pending(all, []int{1, 2, 4})
	if !errors.Is(err, ErrUnknownVersion) {
		t.Errorf("Pending with an applied version 4 that has no file: err = %v, want ErrUnknownVersion", err)
	}
	if err != nil && !strings.Contains(err.Error(), "4") {
		t.Errorf("error %q doesn't say which version", err)
	}

	// 3 was merged after 5 had already run in this database.
	_, err = Pending(all, []int{1, 2, 5})
	if err == nil {
		t.Fatal("Pending([1 2 5]) returned no error, want one: 3 is older than 5, which is applied")
	}
	if errors.Is(err, ErrUnknownVersion) {
		t.Errorf("err = %v; this isn't an unknown version, it's one out of order", err)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("error %q doesn't say which version", err)
	}
}

func TestToRollBack(t *testing.T) {
	tests := []struct {
		name    string
		applied []int
		steps   int
		want    []int
	}{
		{"one step", []int{1, 2, 3}, 1, []int{3}},
		{"two steps, newest first", []int{1, 2, 3}, 2, []int{3, 2}},
		{"more steps than applied", []int{1, 2}, 10, []int{2, 1}},
		{"table order doesn't matter", []int{3, 1, 5, 2}, 2, []int{5, 3}},
		{"nothing applied", nil, 1, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToRollBack(all, tt.applied, tt.steps)
			if err != nil {
				t.Fatalf("ToRollBack(%v, %d) error = %v", tt.applied, tt.steps, err)
			}
			if g := versions(got); !reflect.DeepEqual(g, tt.want) {
				t.Errorf("ToRollBack(%v, %d) = %v, want %v", tt.applied, tt.steps, g, tt.want)
			}
		})
	}
	if _, err := ToRollBack(all, []int{1}, 0); err == nil {
		t.Error("ToRollBack with 0 steps returned no error")
	}
	if _, err := ToRollBack(all, []int{1, 9}, 1); !errors.Is(err, ErrUnknownVersion) {
		t.Errorf("ToRollBack with an unknown applied version: err = %v, want ErrUnknownVersion", err)
	}
}
