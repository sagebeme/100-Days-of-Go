package matatufare

import "testing"

func TestFare(t *testing.T) {
	tests := []struct {
		distanceKm int
		hour       int
		raining    bool
		want       int
	}{
		{3, 10, false, 50},
		{12, 7, false, 110},
		{12, 7, true, 130},
		{40, 18, true, 250},
		{1, 12, false, 50},
		{5, 9, false, 50},
		{6, 5, false, 80},
		{15, 17, false, 110},
		{16, 20, true, 140},
		{30, 19, false, 150},
		{31, 0, false, 200},
		{25, 6, false, 150},
		{25, 8, false, 150},
		{25, 16, false, 120},
		{2, 23, true, 70},
		{0, 7, true, 0},
		{-4, 18, false, 0},
	}
	for _, tt := range tests {
		got := Fare(tt.distanceKm, tt.hour, tt.raining)
		if got != tt.want {
			t.Errorf("Fare(%d, %d, %t)\n got: %d\nwant: %d", tt.distanceKm, tt.hour, tt.raining, got, tt.want)
		}
	}
}

func TestRoute(t *testing.T) {
	tests := []struct {
		destination string
		want        string
	}{
		{"Kawangware", "46"},
		{"Ngong", "111"},
		{"Rongai", "125"},
		{"Ongata Rongai", "125"},
		{"Githurai", "45"},
		{"Buruburu", "58"},
		{"Mombasa", "unknown"},
		{"rongai", "unknown"},
		{"", "unknown"},
	}
	for _, tt := range tests {
		got := Route(tt.destination)
		if got != tt.want {
			t.Errorf("Route(%q)\n got: %q\nwant: %q", tt.destination, got, tt.want)
		}
	}
}
