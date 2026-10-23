package bundlecalculator

import "testing"

func TestBundlesYouCanBuy(t *testing.T) {
	tests := []struct {
		budget, price int
		want          int
	}{
		{250, 99, 2},
		{1000, 20, 50},
		{99, 99, 1},
		{50, 99, 0},
		{0, 99, 0},
	}
	for _, tt := range tests {
		got := BundlesYouCanBuy(tt.budget, tt.price)
		if got != tt.want {
			t.Errorf("BundlesYouCanBuy(%d, %d)\n got: %d\nwant: %d", tt.budget, tt.price, got, tt.want)
		}
	}
}

func TestChange(t *testing.T) {
	tests := []struct {
		budget, price int
		want          int
	}{
		{250, 99, 52},
		{1000, 20, 0},
		{99, 99, 0},
		{50, 99, 50},
		{1001, 250, 1},
	}
	for _, tt := range tests {
		got := Change(tt.budget, tt.price)
		if got != tt.want {
			t.Errorf("Change(%d, %d)\n got: %d\nwant: %d", tt.budget, tt.price, got, tt.want)
		}
	}
}

func TestGBToMB(t *testing.T) {
	tests := []struct {
		gb   float64
		want int
	}{
		{1.5, 1536},
		{2, 2048},
		{0.5, 512},
		{0.3, 307},
		{0, 0},
	}
	for _, tt := range tests {
		got := GBToMB(tt.gb)
		if got != tt.want {
			t.Errorf("GBToMB(%v)\n got: %d\nwant: %d", tt.gb, got, tt.want)
		}
	}
}

func TestPricePerGB(t *testing.T) {
	tests := []struct {
		price, mb int
		want      float64
	}{
		{99, 1024, 99},
		{50, 350, 146.29},
		{20, 250, 81.92},
		{1000, 10240, 100},
		{5, 50, 102.4},
	}
	for _, tt := range tests {
		got := PricePerGB(tt.price, tt.mb)
		if got != tt.want {
			t.Errorf("PricePerGB(%d, %d)\n got: %v\nwant: %v", tt.price, tt.mb, got, tt.want)
		}
	}
}
