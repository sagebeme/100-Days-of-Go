package graphs

import (
	"slices"
	"strconv"
	"testing"
)

// nairobi builds a small map of real Nairobi stops. The connections are
// simplified for the exercise: one link per route listed on nishukishe.com
// (checked October 2026), plus short walks between CBD stages.
// Routes change, so confirm with the conductor.
func nairobi() *Graph {
	var g Graph
	// Walking links inside the CBD (simplified).
	g.AddRoute("Kencom", "Ambassadeur")
	g.AddRoute("Ambassadeur", "Ronald Ngala Street")
	g.AddRoute("Kencom", "Railways")
	// One link per matatu route.
	g.AddRoute("Kencom", "Kawangware")            // 46K, Citi Hoppa
	g.AddRoute("Railways", "Ngong")               // 111, MCA Sacco
	g.AddRoute("Railways", "Ongata Rongai")       // 125, City Shuttle
	g.AddRoute("Ronald Ngala Street", "Githurai") // 45, Nicco Travellers
	g.AddRoute("Ambassadeur", "Buruburu")         // 58, Buruburu 58 Travellers Sacco
	g.AddRoute("Railways", "Kibera")              // Highrise Kibera Sacco
	return &g
}

// villages is a made-up network with two separate parts.
func villages() *Graph {
	var g Graph
	for _, r := range [][2]string{
		{"Amani", "Baraka"}, {"Baraka", "Chui"}, {"Amani", "Dawa"}, {"Dawa", "Chui"}, {"Chui", "Embe"},
		{"Fahari", "Giza"},
	} {
		g.AddRoute(r[0], r[1])
	}
	return &g
}

func TestAddRouteAndNeighbours(t *testing.T) {
	g := nairobi()
	tests := []struct {
		stop string
		want []string
	}{
		{"Kencom", []string{"Ambassadeur", "Kawangware", "Railways"}},
		{"Railways", []string{"Kencom", "Kibera", "Ngong", "Ongata Rongai"}},
		{"Githurai", []string{"Ronald Ngala Street"}},
		{"Westlands", nil},
	}
	for _, tt := range tests {
		if got := g.Neighbours(tt.stop); !slices.Equal(got, tt.want) {
			t.Errorf("Neighbours(%q) = %q, want %q", tt.stop, got, tt.want)
		}
	}

	var h Graph
	h.AddRoute("Amani", "Baraka")
	h.AddRoute("Baraka", "Amani") // the same connection again
	h.AddRoute("Amani", "Amani")  // a loop to itself is ignored
	if got := h.Neighbours("Amani"); !slices.Equal(got, []string{"Baraka"}) {
		t.Errorf("after duplicate routes, Neighbours(Amani) = %q, want [Baraka]", got)
	}
}

func TestFewestHops(t *testing.T) {
	g := nairobi()
	tests := []struct {
		from, to string
		want     int
		ok       bool
	}{
		{"Kencom", "Kencom", 0, true},
		{"Kencom", "Kawangware", 1, true},
		{"Kencom", "Ngong", 2, true},
		{"Githurai", "Ngong", 5, true},
		{"Buruburu", "Kibera", 4, true},
		{"Kencom", "Westlands", 0, false},
		{"Westlands", "Westlands", 0, false},
	}
	for _, tt := range tests {
		got, ok := g.FewestHops(tt.from, tt.to)
		if got != tt.want || ok != tt.ok {
			t.Errorf("FewestHops(%q, %q) = %d, %v, want %d, %v", tt.from, tt.to, got, ok, tt.want, tt.ok)
		}
	}
	v := villages()
	if got, ok := v.FewestHops("Amani", "Embe"); got != 3 || !ok {
		t.Errorf("villages FewestHops(Amani, Embe) = %d, %v, want 3, true", got, ok)
	}
	if _, ok := v.FewestHops("Amani", "Giza"); ok {
		t.Errorf("villages FewestHops(Amani, Giza) found a route between separate networks")
	}
}

func TestPath(t *testing.T) {
	g := nairobi()
	tests := []struct {
		from, to string
		want     []string
	}{
		{"Kencom", "Kencom", []string{"Kencom"}},
		{"Githurai", "Ngong", []string{"Githurai", "Ronald Ngala Street", "Ambassadeur", "Kencom", "Railways", "Ngong"}},
		{"Kibera", "Buruburu", []string{"Kibera", "Railways", "Kencom", "Ambassadeur", "Buruburu"}},
		{"Kencom", "Westlands", nil},
	}
	for _, tt := range tests {
		if got := g.Path(tt.from, tt.to); !slices.Equal(got, tt.want) {
			t.Errorf("Path(%q, %q)\n got: %q\nwant: %q", tt.from, tt.to, got, tt.want)
		}
	}
	// Two routes of the same length: Amani-Baraka-Chui and Amani-Dawa-Chui.
	// Visiting neighbours alphabetically, BFS reaches Chui through Baraka first.
	v := villages()
	if got, want := v.Path("Amani", "Embe"), []string{"Amani", "Baraka", "Chui", "Embe"}; !slices.Equal(got, want) {
		t.Errorf("villages Path(Amani, Embe) = %q, want %q", got, want)
	}
}

func TestReachable(t *testing.T) {
	v := villages()
	tests := []struct {
		start string
		want  []string
	}{
		{"Amani", []string{"Amani", "Baraka", "Chui", "Dawa", "Embe"}},
		{"Embe", []string{"Amani", "Baraka", "Chui", "Dawa", "Embe"}},
		{"Giza", []string{"Fahari", "Giza"}},
		{"Nowhere", []string{}},
	}
	for _, tt := range tests {
		got := v.Reachable(tt.start)
		if got == nil || !slices.Equal(got, tt.want) {
			t.Errorf("Reachable(%q) = %#v, want %q", tt.start, got, tt.want)
		}
	}
	if got := nairobi().Reachable("Kibera"); len(got) != 10 {
		t.Errorf("nairobi Reachable(Kibera) has %d stops, want all 10: %q", len(got), got)
	}
}

func TestCountNetworks(t *testing.T) {
	var empty Graph
	if got := empty.CountNetworks(); got != 0 {
		t.Errorf("empty graph CountNetworks() = %d, want 0", got)
	}
	if got := nairobi().CountNetworks(); got != 1 {
		t.Errorf("nairobi CountNetworks() = %d, want 1", got)
	}
	if got := villages().CountNetworks(); got != 2 {
		t.Errorf("villages CountNetworks() = %d, want 2", got)
	}
}

func TestBigLine(t *testing.T) {
	// 100 000 stops in a line, plus 100 separate pairs. BFS and DFS are
	// O(V + E); trying every route would never finish.
	var g Graph
	for i := 1; i < 100_000; i++ {
		g.AddRoute("s"+strconv.Itoa(i-1), "s"+strconv.Itoa(i))
	}
	for i := range 100 {
		g.AddRoute("p"+strconv.Itoa(i), "q"+strconv.Itoa(i))
	}
	if got, ok := g.FewestHops("s0", "s99999"); got != 99_999 || !ok {
		t.Errorf("FewestHops along the line = %d, %v, want 99999, true", got, ok)
	}
	if got := len(g.Reachable("s500")); got != 100_000 {
		t.Errorf("len(Reachable(s500)) = %d, want 100000", got)
	}
	if got := g.CountNetworks(); got != 101 {
		t.Errorf("CountNetworks() = %d, want 101", got)
	}
}
