package bodafleet

import (
	"reflect"
	"testing"
)

func TestLabel(t *testing.T) {
	tests := []struct {
		bike Bike
		want string
	}{
		{Bike{Plate: "KMFA 123B", Rider: "Otieno"}, "KMFA 123B (Otieno)"},
		{Bike{Plate: "KMEX 777Q", Rider: "Wanjiru Kamau"}, "KMEX 777Q (Wanjiru Kamau)"},
		{Bike{Plate: "KMGG 404Z"}, "KMGG 404Z (no rider)"},
	}
	for _, tt := range tests {
		got := tt.bike.Label()
		if got != tt.want {
			t.Errorf("%+v.Label()\n got: %q\nwant: %q", tt.bike, got, tt.want)
		}
	}
}

func TestRide(t *testing.T) {
	b := Bike{Plate: "KMFA 123B", Rider: "Otieno", Odometer: 5000, Trips: 10, Earnings: 2000}
	if err := b.Ride(4, 150); err != nil {
		t.Fatalf("Ride(4, 150) returned an error: %v", err)
	}
	if err := b.Ride(12, 400); err != nil {
		t.Fatalf("Ride(12, 400) returned an error: %v", err)
	}
	want := Bike{Plate: "KMFA 123B", Rider: "Otieno", Odometer: 5016, Trips: 12, Earnings: 2550}
	if b != want {
		t.Errorf("after Ride(4, 150) and Ride(12, 400)\n got: %+v\nwant: %+v", b, want)
	}
}

func TestRideRejects(t *testing.T) {
	tests := []struct {
		km, fare int
	}{
		{0, 100},
		{-3, 100},
		{5, -50},
	}
	for _, tt := range tests {
		b := Bike{Plate: "KMFA 123B", Odometer: 100, Trips: 1, Earnings: 70}
		before := b
		err := b.Ride(tt.km, tt.fare)
		if err == nil {
			t.Errorf("Ride(%d, %d) = nil, want an error", tt.km, tt.fare)
		}
		if b != before {
			t.Errorf("Ride(%d, %d) changed the bike on an error\n got: %+v\nwant: %+v", tt.km, tt.fare, b, before)
		}
	}
}

func TestNeedsServiceAndService(t *testing.T) {
	tests := []struct {
		odometer, last int
		want           bool
	}{
		{0, 0, false},
		{999, 0, false},
		{1000, 0, true},
		{5400, 4300, true},
		{5400, 4401, false},
	}
	for _, tt := range tests {
		b := Bike{Plate: "KMFA 123B", Odometer: tt.odometer, LastServiceKm: tt.last}
		if got := b.NeedsService(); got != tt.want {
			t.Errorf("Bike{Odometer: %d, LastServiceKm: %d}.NeedsService() = %v, want %v", tt.odometer, tt.last, got, tt.want)
		}
	}

	b := Bike{Plate: "KMFA 123B", Odometer: 2300, LastServiceKm: 1000}
	b.Service()
	if b.LastServiceKm != 2300 || b.NeedsService() {
		t.Errorf("after Service(), LastServiceKm = %d and NeedsService() = %v; want 2300 and false", b.LastServiceKm, b.NeedsService())
	}
}

func newFleet(t *testing.T) *Fleet {
	t.Helper()
	f := &Fleet{}
	bikes := []Bike{
		{Plate: "KMFA 123B", Rider: "Otieno", Odometer: 990, Earnings: 1200},
		{Plate: "KMDD 456C", Rider: "Achieng", Odometer: 2500, LastServiceKm: 2000, Earnings: 800},
		{Plate: "KMCC 789D", Odometer: 3100, LastServiceKm: 2000},
	}
	for _, b := range bikes {
		if err := f.Add(b); err != nil {
			t.Fatalf("Add(%+v) returned an error: %v", b, err)
		}
	}
	return f
}

func TestFleetAdd(t *testing.T) {
	f := newFleet(t)
	if len(f.Bikes) != 3 {
		t.Fatalf("after adding 3 bikes, len(f.Bikes) = %d, want 3", len(f.Bikes))
	}
	if err := f.Add(Bike{Plate: "KMFA 123B", Rider: "Someone else"}); err == nil {
		t.Errorf("Add with a plate that's already there = nil, want an error")
	}
	if err := f.Add(Bike{Rider: "Mwangi"}); err == nil {
		t.Errorf("Add with an empty plate = nil, want an error")
	}
	if len(f.Bikes) != 3 {
		t.Errorf("after two rejected Adds, len(f.Bikes) = %d, want 3", len(f.Bikes))
	}
}

func TestFleetFind(t *testing.T) {
	f := newFleet(t)
	b := f.Find("KMDD 456C")
	if b == nil {
		t.Fatalf("Find(%q) = nil, want the bike", "KMDD 456C")
	}
	if b.Rider != "Achieng" {
		t.Errorf("Find(%q).Rider = %q, want %q", "KMDD 456C", b.Rider, "Achieng")
	}
	b.Rider = "Njoroge"
	if got := f.Find("KMDD 456C").Rider; got != "Njoroge" {
		t.Errorf("after changing the bike Find returned, the fleet's bike has Rider %q, want %q (return &f.Bikes[i], not a copy)", got, "Njoroge")
	}
	if got := f.Find("KAA 000A"); got != nil {
		t.Errorf("Find(%q) = %+v, want nil", "KAA 000A", got)
	}
}

func TestFleetRide(t *testing.T) {
	f := newFleet(t)
	if err := f.Ride("KMFA 123B", 15, 300); err != nil {
		t.Fatalf("Ride(%q, 15, 300) returned an error: %v", "KMFA 123B", err)
	}
	b := f.Find("KMFA 123B")
	if b == nil {
		t.Fatalf("Find(%q) = nil after Add, want the bike", "KMFA 123B")
	}
	if b.Odometer != 1005 || b.Earnings != 1500 || b.Trips != 1 {
		t.Errorf("after Ride, the bike is %+v; want Odometer 1005, Earnings 1500, Trips 1", *b)
	}
	if err := f.Ride("KAA 000A", 3, 100); err == nil {
		t.Errorf("Ride on a plate that isn't in the fleet = nil, want an error")
	}
	if err := f.Ride("KMFA 123B", 0, 100); err == nil {
		t.Errorf("Ride with 0 km = nil, want an error")
	}
}

func TestTotalEarnings(t *testing.T) {
	f := newFleet(t)
	if got := f.TotalEarnings(); got != 2000 {
		t.Errorf("TotalEarnings() = %d, want 2000", got)
	}
	if err := f.Ride("KMCC 789D", 7, 250); err != nil {
		t.Fatalf("Ride returned an error: %v", err)
	}
	if got := f.TotalEarnings(); got != 2250 {
		t.Errorf("after a Ksh250 trip, TotalEarnings() = %d, want 2250", got)
	}
	if got := (Fleet{}).TotalEarnings(); got != 0 {
		t.Errorf("Fleet{}.TotalEarnings() = %d, want 0", got)
	}
}

func TestDueForService(t *testing.T) {
	f := newFleet(t)
	if got, want := f.DueForService(), []string{"KMCC 789D"}; !reflect.DeepEqual(got, want) {
		t.Errorf("DueForService()\n got: %q\nwant: %q", got, want)
	}

	// KMFA 123B passes 1000 km; KMCC 789D gets serviced.
	if err := f.Ride("KMFA 123B", 20, 500); err != nil {
		t.Fatalf("Ride returned an error: %v", err)
	}
	cc := f.Find("KMCC 789D")
	if cc == nil {
		t.Fatalf("Find(%q) = nil after Add, want the bike", "KMCC 789D")
	}
	cc.Service()
	if err := f.Ride("KMDD 456C", 600, 9000); err != nil {
		t.Fatalf("Ride returned an error: %v", err)
	}
	if got, want := f.DueForService(), []string{"KMDD 456C", "KMFA 123B"}; !reflect.DeepEqual(got, want) {
		t.Errorf("DueForService() after more rides and a service\n got: %q\nwant: %q", got, want)
	}

	if got := (Fleet{}).DueForService(); got == nil || len(got) != 0 {
		t.Errorf("Fleet{}.DueForService() = %#v, want []string{}", got)
	}
}
