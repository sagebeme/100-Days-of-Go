package threatmodelling

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

var small = Model{
	Components: []Component{
		{Name: "Conductor", Kind: External, Zone: "internet"},
		{Name: "API", Kind: Process, Zone: "dmz"},
		{Name: "DB", Kind: DataStore, Zone: "internal"},
	},
	Flows: []Flow{
		{From: "Conductor", To: "API", Data: "check-ins", Encrypted: true, Authenticated: true},
		{From: "API", To: "DB", Data: "bookings", Encrypted: true, Authenticated: false},
	},
}

func TestValidate(t *testing.T) {
	if err := small.Validate(); err != nil {
		t.Errorf("Validate(small) = %v, want nil", err)
	}
	if err := TwendeModel().Validate(); err != nil {
		t.Errorf("Validate(TwendeModel) = %v, want nil", err)
	}
	tests := []struct {
		name string
		m    Model
		want error
		who  string
	}{
		{"duplicate", Model{Components: []Component{{Name: "API", Kind: Process}, {Name: "API", Kind: DataStore}}}, ErrDuplicate, "API"},
		{"bad kind", Model{Components: []Component{{Name: "Kafka", Kind: "queue"}}}, ErrBadKind, "Kafka"},
		{"unknown from", Model{Components: []Component{{Name: "API", Kind: Process}}, Flows: []Flow{{From: "Ghost", To: "API"}}}, ErrUnknown, "Ghost"},
		{"unknown to", Model{Components: []Component{{Name: "API", Kind: Process}}, Flows: []Flow{{From: "API", To: "S3"}}}, ErrUnknown, "S3"},
	}
	for _, tt := range tests {
		err := tt.m.Validate()
		if !errors.Is(err, tt.want) {
			t.Errorf("%s: Validate = %v, want %v", tt.name, err, tt.want)
			continue
		}
		if !strings.Contains(err.Error(), tt.who) {
			t.Errorf("%s: error %q doesn't name %q", tt.name, err, tt.who)
		}
	}
}

func TestCrossesBoundary(t *testing.T) {
	m := TwendeModel()
	want := []bool{true, true, true, true, true, false} // only Seat service -> Postgres stays inside one zone
	for i, f := range m.Flows {
		if got := m.CrossesBoundary(f); got != want[i] {
			t.Errorf("CrossesBoundary(%s -> %s) = %v, want %v", f.From, f.To, got, want[i])
		}
	}
}

// brief turns threats into "Target: Category" strings for easy comparison.
func brief(ts []Threat) []string {
	var out []string
	for _, t := range ts {
		out = append(out, t.Target+": "+string(t.Category))
	}
	return out
}

func TestSTRIDESmall(t *testing.T) {
	got, err := STRIDE(small)
	if err != nil {
		t.Fatalf("STRIDE(small) = %v", err)
	}
	want := []string{
		"Conductor: Spoofing", "Conductor: Repudiation",
		"API: Spoofing", "API: Tampering", "API: Repudiation", "API: Information disclosure", "API: Denial of service", "API: Elevation of privilege",
		"DB: Tampering", "DB: Information disclosure", "DB: Denial of service",
		"Conductor -> API: Tampering", "Conductor -> API: Information disclosure", "Conductor -> API: Denial of service",
		"API -> DB: Spoofing", "API -> DB: Tampering", "API -> DB: Information disclosure", "API -> DB: Denial of service",
	}
	if b := brief(got); !slices.Equal(b, want) {
		t.Errorf("STRIDE(small) =\n%s\nwant\n%s", strings.Join(b, "\n"), strings.Join(want, "\n"))
	}
	for _, th := range got {
		if q := fmt.Sprintf(Questions[th.Category], th.Target); th.Question != q {
			t.Errorf("%s %s: Question = %q, want %q", th.Target, th.Category, th.Question, q)
		}
		isFlow := strings.Contains(th.Target, " -> ")
		if th.Boundary != isFlow {
			t.Errorf("%s: Boundary = %v, want %v (both flows in small cross a boundary; components never do)", th.Target, th.Boundary, isFlow)
		}
	}
}

func TestSTRIDETwende(t *testing.T) {
	got, err := STRIDE(TwendeModel())
	if err != nil {
		t.Fatalf("STRIDE(Twende) = %v", err)
	}
	// 2 externals × 2 + 2 processes × 6 + 2 stores × 3 = 22, then 6 flows × 3 = 18,
	// plus Spoofing for the two unauthenticated flows that cross a boundary.
	if len(got) != 42 {
		t.Errorf("STRIDE(Twende) gave %d threats, want 42", len(got))
	}
	spoofable := map[string]bool{}
	for _, th := range got {
		if th.Category == Spoofing && strings.Contains(th.Target, " -> ") {
			spoofable[th.Target] = true
		}
	}
	for _, want := range []string{"M-Pesa -> Booking API", "Booking API -> Seat service"} {
		if !spoofable[want] {
			t.Errorf("no Spoofing threat for %q: it crosses a boundary and isn't authenticated", want)
		}
	}
	if len(spoofable) != 2 {
		t.Errorf("Spoofing flagged for flows %v, want only the 2 unauthenticated boundary crossings", spoofable)
	}
}

func TestSTRIDEInvalid(t *testing.T) {
	bad := Model{Flows: []Flow{{From: "a", To: "b"}}}
	if got, err := STRIDE(bad); !errors.Is(err, ErrUnknown) || got != nil {
		t.Errorf("STRIDE(invalid) = %v, %v, want nil, ErrUnknown", got, err)
	}
}

func TestScoreAndLevel(t *testing.T) {
	tests := []struct {
		l, i  int
		score int
		level string
	}{
		{1, 1, 1, "low"},
		{2, 2, 4, "low"},
		{1, 5, 5, "medium"},
		{3, 3, 9, "medium"},
		{2, 5, 10, "high"},
		{4, 4, 16, "high"},
		{4, 5, 20, "critical"},
		{5, 5, 25, "critical"},
	}
	for _, tt := range tests {
		s, err := Score(tt.l, tt.i)
		if err != nil || s != tt.score {
			t.Errorf("Score(%d, %d) = %d, %v, want %d, nil", tt.l, tt.i, s, err, tt.score)
		}
		if got := Level(tt.score); got != tt.level {
			t.Errorf("Level(%d) = %q, want %q", tt.score, got, tt.level)
		}
	}
	for _, bad := range [][2]int{{0, 3}, {3, 0}, {6, 1}, {1, 6}, {-1, -1}} {
		if _, err := Score(bad[0], bad[1]); !errors.Is(err, ErrScale) {
			t.Errorf("Score(%d, %d) error = %v, want ErrScale", bad[0], bad[1], err)
		}
	}
}

func TestPrioritise(t *testing.T) {
	r := func(cat Category, target string, l, i int) Risk {
		return Risk{Threat: Threat{Category: cat, Target: target}, Likelihood: l, Impact: i}
	}
	in := []Risk{
		r(DenialOfService, "Booking API", 4, 2),                // 8
		r(Spoofing, "M-Pesa -> Booking API", 4, 5),             // 20
		r(InformationDisclosure, "Postgres", 2, 4),             // 8, but impact 4 beats impact 2
		r(Tampering, "Redis", 2, 4),                            // 8, impact 4, Tampering before Information disclosure
		r(Repudiation, "Passenger app", 5, 4),                  // 20, impact 4 loses to impact 5
		r(ElevationOfPrivilege, "Booking API", 2, 4),           // 8, impact 4, category last
		r(InformationDisclosure, "Booking API -> Redis", 2, 4), // same as Postgres but "B" < "P"
	}
	before := slices.Clone(in)
	got, err := Prioritise(in)
	if err != nil {
		t.Fatalf("Prioritise = %v", err)
	}
	want := []string{
		"M-Pesa -> Booking API", "Passenger app",
		"Redis", "Booking API -> Redis", "Postgres", "Booking API",
		"Booking API",
	}
	var targets []string
	for _, x := range got {
		targets = append(targets, x.Threat.Target)
	}
	if !slices.Equal(targets, want) {
		t.Errorf("Prioritise order = %q, want %q", targets, want)
	}
	if got[5].Threat.Category != ElevationOfPrivilege || got[6].Threat.Category != DenialOfService {
		t.Errorf("the two Booking API risks are in the wrong order: %v then %v", got[5].Threat.Category, got[6].Threat.Category)
	}
	if !slices.Equal(in, before) {
		t.Errorf("Prioritise changed its input; sort a copy")
	}

	_, err = Prioritise([]Risk{r(Spoofing, "Redis", 3, 9)})
	if !errors.Is(err, ErrScale) || !strings.Contains(err.Error(), "Redis") {
		t.Errorf("Prioritise(bad rating) = %v, want ErrScale naming Redis", err)
	}
}
