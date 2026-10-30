package routes

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewNetwork(t *testing.T) {
	bad := [][]Link{
		{{"A", "", "1", 5}},
		{{"A", "A", "1", 5}},
		{{"A", "B", "", 5}},
		{{"A", "B", "1", 0}},
		{{"A", "B", "1", -3}},
	}
	for _, links := range bad {
		if _, err := NewNetwork(links); err == nil {
			t.Errorf("NewNetwork(%v) succeeded, want an error", links)
		}
	}
	n, err := NewNetwork([]Link{{"Kencom", "Railways", Walk, 8}, {"Railways", "Ngong", "111", 70}})
	if err != nil {
		t.Fatalf("NewNetwork: %v", err)
	}
	if got, want := n.Stops(), []string{"Kencom", "Railways", "Ngong"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Stops() = %v, want %v", got, want)
	}
	if !n.Has("Ngong") || n.Has("Westlands") {
		t.Errorf("Has: want Ngong true and Westlands false")
	}
}

func TestNairobiFastest(t *testing.T) {
	n := Nairobi()
	tests := []struct {
		from, to string
		want     Journey
	}{
		{"Kawangware", "Ngong", Journey{
			Stops: []string{"Kawangware", "Kencom", "Railways", "Ngong"},
			Legs: []Leg{
				{"46K", "Kawangware", "Kencom", 45},
				{Walk, "Kencom", "Railways", 8},
				{"111", "Railways", "Ngong", 70},
			},
			Minutes: 123, Transfers: 1,
		}},
		{"Buruburu", "Githurai", Journey{
			Stops: []string{"Buruburu", "Ambassadeur", "Ronald Ngala Street", "Githurai"},
			Legs: []Leg{
				{"58", "Buruburu", "Ambassadeur", 35},
				{Walk, "Ambassadeur", "Ronald Ngala Street", 6},
				{"45", "Ronald Ngala Street", "Githurai", 50},
			},
			Minutes: 91, Transfers: 1,
		}},
		{"Kencom", "Kibera", Journey{
			Stops:   []string{"Kencom", "Railways", "Kibera"},
			Legs:    []Leg{{Walk, "Kencom", "Railways", 8}, {"Highrise Kibera Sacco", "Railways", "Kibera", 30}},
			Minutes: 38, Transfers: 0,
		}},
		{"Kencom", "Ronald Ngala Street", Journey{
			Stops:   []string{"Kencom", "Ambassadeur", "Ronald Ngala Street"},
			Legs:    []Leg{{Walk, "Kencom", "Ronald Ngala Street", 9}},
			Minutes: 9, Transfers: 0,
		}},
		{"Ngong", "Ngong", Journey{Stops: []string{"Ngong"}}},
	}
	for _, tt := range tests {
		got, err := n.Fastest(tt.from, tt.to)
		if err != nil {
			t.Errorf("Fastest(%s, %s): %v", tt.from, tt.to, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Fastest(%s, %s) =\n%+v\nwant\n%+v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestUnknownStop(t *testing.T) {
	n := Nairobi()
	if _, err := n.Fastest("Westlands", "Ngong"); !errors.Is(err, ErrUnknownStop) {
		t.Errorf("Fastest from an unknown stop: %v, want ErrUnknownStop", err)
	}
	if _, err := n.FewestTransfers("Ngong", "Westlands"); !errors.Is(err, ErrUnknownStop) {
		t.Errorf("FewestTransfers to an unknown stop: %v, want ErrUnknownStop", err)
	}
}

// A made-up network, not real places: Mto, Soko, Shule and Ziwa.
var madeUp = []Link{
	{"Mto", "Soko", "A", 10},
	{"Soko", "Ziwa", "B", 10},
	{"Mto", "Shule", "C", 15},
	{"Shule", "Ziwa", "C", 15},
	{"Mto", "Kona", "D", 5},
	{"Kona", "Pembe", Walk, 2},
	{"Pembe", "Mwisho", "D", 5},
	{"Kisiwa", "Pwani", "E", 5},
	{"X", "Y", "P", 10},
	{"Y", "Z", "Q", 10},
	{"X", "Z", "R", 20},
}

func TestFewestTransfers(t *testing.T) {
	n, err := NewNetwork(madeUp)
	if err != nil {
		t.Fatal(err)
	}

	fast, _ := n.Fastest("Mto", "Ziwa")
	if fast.Minutes != 20 || fast.Transfers != 1 {
		t.Errorf("Fastest(Mto, Ziwa) = %d min, %d transfers, want 20 min, 1 transfer", fast.Minutes, fast.Transfers)
	}

	got, err := n.FewestTransfers("Mto", "Ziwa")
	if err != nil {
		t.Fatalf("FewestTransfers: %v", err)
	}
	want := Journey{
		Stops:   []string{"Mto", "Shule", "Ziwa"},
		Legs:    []Leg{{"C", "Mto", "Ziwa", 30}},
		Minutes: 30, Transfers: 0,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FewestTransfers(Mto, Ziwa) =\n%+v\nwant\n%+v", got, want)
	}
}

func TestWalkingBetweenTwoRidesIsATransfer(t *testing.T) {
	n, _ := NewNetwork(madeUp)
	j, err := n.FewestTransfers("Mto", "Mwisho")
	if err != nil {
		t.Fatal(err)
	}
	if len(j.Legs) != 3 || j.Transfers != 1 {
		t.Errorf("Mto to Mwisho = %d legs, %d transfers, want 3 legs (D, walk, D) and 1 transfer", len(j.Legs), j.Transfers)
	}
}

func TestFastestPrefersFewerTransfersOnATie(t *testing.T) {
	n, _ := NewNetwork(madeUp)
	j, _ := n.Fastest("X", "Z")
	if j.Minutes != 20 || j.Transfers != 0 || len(j.Legs) != 1 || j.Legs[0].Route != "R" {
		t.Errorf("Fastest(X, Z) = %+v, want the direct R in 20 min", j)
	}
}

func TestNoJourney(t *testing.T) {
	n, _ := NewNetwork(madeUp)
	for _, find := range []func(string, string) (Journey, error){n.Fastest, n.FewestTransfers} {
		if _, err := find("Mto", "Pwani"); !errors.Is(err, ErrNoJourney) {
			t.Errorf("Mto to Pwani: %v, want ErrNoJourney", err)
		}
	}
}

func TestDescribe(t *testing.T) {
	j, _ := Nairobi().Fastest("Kawangware", "Ngong")
	want := "Take the 46K from Kawangware to Kencom (45 min)\n" +
		"Walk from Kencom to Railways (8 min)\n" +
		"Take the 111 from Railways to Ngong (70 min)\n" +
		"Total: 123 min, 1 transfer"
	if got := Describe(j); got != want {
		t.Errorf("Describe =\n%s\nwant\n%s", got, want)
	}
	if got := Describe(Journey{Stops: []string{"Ngong"}}); got != "Total: 0 min, 0 transfers" {
		t.Errorf("Describe(empty journey) = %q", got)
	}
}

func TestNairobiData(t *testing.T) {
	n := Nairobi()
	for _, stop := range []string{"Kencom", "Ambassadeur", "Railways", "Ronald Ngala Street",
		"Kawangware", "Ngong", "Ongata Rongai", "Githurai", "Buruburu", "Kibera"} {
		if !n.Has(stop) {
			t.Errorf("the Nairobi network has no %s", stop)
		}
	}
	if got := len(n.Stops()); got != 10 {
		t.Errorf("the Nairobi network has %d stops, want 10 (only places checked on nishukishe.com)", got)
	}
}
