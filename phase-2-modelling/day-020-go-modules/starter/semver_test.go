package gomodules

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Version
	}{
		{"v1.2.3", Version{1, 2, 3, ""}},
		{"v0.0.0", Version{0, 0, 0, ""}},
		{"v10.20.30", Version{10, 20, 30, ""}},
		{"v2.0.0-rc.1", Version{2, 0, 0, "rc.1"}},
		{"v1.0.0-alpha", Version{1, 0, 0, "alpha"}},
		{"v1.0.0-alpha.beta.7", Version{1, 0, 0, "alpha.beta.7"}},
		{"v1.0.0-x-y-z.0", Version{1, 0, 0, "x-y-z.0"}},
		{"v0.3.1-0.20260301120000-abcdef123456", Version{0, 3, 1, "0.20260301120000-abcdef123456"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil || got != tt.want {
				t.Errorf("Parse(%q)\n got: %+v, %v\nwant: %+v, <nil>", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	bad := []string{
		"",
		"1.2.3",         // no v
		"V1.2.3",        // capital V
		"v1.2",          // two numbers
		"v1.2.3.4",      // four numbers
		"v1..3",         // missing number
		"v1.02.3",       // leading zero
		"v01.2.3",       // leading zero
		"v1.2.x",        // not a number
		"v-1.2.3",       // sign
		"v1.2.3-",       // empty pre-release
		"v1.2.3-rc..1",  // empty identifier
		"v1.2.3-rc.01",  // numeric identifier with a leading zero
		"v1.2.3-rc_1",   // bad character
		"v1.2.3+build5", // build metadata
		"v 1.2.3",
	}
	for _, s := range bad {
		t.Run(s, func(t *testing.T) {
			if got, err := Parse(s); err == nil {
				t.Errorf("Parse(%q) = %+v, nil; want an error", s, got)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{Version{1, 2, 3, ""}, "v1.2.3"},
		{Version{0, 10, 0, ""}, "v0.10.0"},
		{Version{2, 0, 0, "rc.1"}, "v2.0.0-rc.1"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("%+v.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q) returned an error: %v", s, err)
	}
	return v
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "v1.2.3", 0},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.3.0", "v1.2.9", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"v1.10.0", "v1.9.0", 1}, // numbers, not strings
		{"v0.9.0", "v1.0.0", -1},
		{"v1.0.0-rc.1", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-rc.1", 1},
		{"v1.0.1-alpha", "v1.0.0", 1},
		{"v1.0.0-rc.1", "v1.0.0-rc.1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			if got := Compare(mustParse(t, tt.a), mustParse(t, tt.b)); got != tt.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestCompareOrder walks the example list from the semver specification, which is in
// increasing order, and checks every pair.
func TestCompareOrder(t *testing.T) {
	order := []string{
		"v1.0.0-alpha",
		"v1.0.0-alpha.1",
		"v1.0.0-alpha.beta",
		"v1.0.0-beta",
		"v1.0.0-beta.2",
		"v1.0.0-beta.11",
		"v1.0.0-rc.1",
		"v1.0.0",
	}
	for i := range order {
		for j := range order {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := Compare(mustParse(t, order[i]), mustParse(t, order[j])); got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", order[i], order[j], got, want)
			}
		}
	}
}

func TestUpgrade(t *testing.T) {
	available := []string{"v1.2.0", "v1.4.1", "v1.3.0", "v1.5.0-rc.1", "v2.0.0", "v2.1.0", "v1.4.0"}
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{"newest v1 release", "v1.2.0", "v1.4.1"},
		{"already newest", "v1.4.1", "v1.4.1"},
		{"never a new major", "v2.0.0", "v2.1.0"},
		{"no older version", "v1.4.5", "v1.4.5"},
		{"pre-release current moves to a release", "v1.4.1-rc.1", "v1.4.1"},
		{"major with nothing available", "v3.0.0", "v3.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Upgrade(tt.current, available)
			if err != nil || got != tt.want {
				t.Errorf("Upgrade(%q, available)\n got: %q, %v\nwant: %q, <nil>", tt.current, got, err, tt.want)
			}
		})
	}

	if got, err := Upgrade("1.2.0", available); err == nil {
		t.Errorf("Upgrade with a bad current version = %q, nil; want an error", got)
	}
	if got, err := Upgrade("v1.2.0", []string{"v1.3.0", "latest"}); err == nil {
		t.Errorf("Upgrade with a bad available version = %q, nil; want an error", got)
	}
}

func TestModulePath(t *testing.T) {
	tests := []struct {
		major int
		want  string
	}{
		{0, "github.com/sagebeme/pesa"},
		{1, "github.com/sagebeme/pesa"},
		{2, "github.com/sagebeme/pesa/v2"},
		{11, "github.com/sagebeme/pesa/v11"},
	}
	for _, tt := range tests {
		if got := ModulePath("github.com/sagebeme/pesa", tt.major); got != tt.want {
			t.Errorf("ModulePath(%q, %d) = %q, want %q", "github.com/sagebeme/pesa", tt.major, got, tt.want)
		}
	}
}
