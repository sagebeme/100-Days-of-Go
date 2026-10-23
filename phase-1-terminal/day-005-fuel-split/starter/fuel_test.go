package fuelsplit

import "testing"

func TestTripCost(t *testing.T) {
	tests := []struct {
		distanceKm, kmPerLitre, pricePerLitre int
		want                                  int
		wantErr                               bool
	}{
		{100, 15, 180, 1200, false},
		{120, 12, 180, 1800, false},
		{50, 12, 177, 738, false},
		{7, 3, 100, 234, false},
		{1, 20, 180, 9, false},
		{0, 10, 180, 0, false},
		{90, 10, 0, 0, false},
		{10, 0, 180, 0, true},
		{10, -5, 180, 0, true},
		{-5, 10, 180, 0, true},
		{10, 10, -180, 0, true},
	}
	for _, tt := range tests {
		got, err := TripCost(tt.distanceKm, tt.kmPerLitre, tt.pricePerLitre)
		if tt.wantErr {
			if err == nil {
				t.Errorf("TripCost(%d, %d, %d) should return an error, got nil", tt.distanceKm, tt.kmPerLitre, tt.pricePerLitre)
			}
			if got != 0 {
				t.Errorf("TripCost(%d, %d, %d) returned an error, so the cost should be 0, got %d", tt.distanceKm, tt.kmPerLitre, tt.pricePerLitre, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("TripCost(%d, %d, %d) returned an unexpected error: %v", tt.distanceKm, tt.kmPerLitre, tt.pricePerLitre, err)
			continue
		}
		if got != tt.want {
			t.Errorf("TripCost(%d, %d, %d)\n got: %d\nwant: %d", tt.distanceKm, tt.kmPerLitre, tt.pricePerLitre, got, tt.want)
		}
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		cost, people int
		wantShare    int
		wantLeftover int
		wantErr      bool
	}{
		{1800, 4, 450, 0, false},
		{738, 4, 184, 2, false},
		{100, 3, 33, 1, false},
		{5, 10, 0, 5, false},
		{999, 1, 999, 0, false},
		{0, 3, 0, 0, false},
		{500, 0, 0, 0, true},
		{500, -2, 0, 0, true},
		{-100, 2, 0, 0, true},
	}
	for _, tt := range tests {
		share, leftover, err := Split(tt.cost, tt.people)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Split(%d, %d) should return an error, got nil", tt.cost, tt.people)
			}
			if share != 0 || leftover != 0 {
				t.Errorf("Split(%d, %d) returned an error, so share and leftover should be 0, got %d and %d", tt.cost, tt.people, share, leftover)
			}
			continue
		}
		if err != nil {
			t.Errorf("Split(%d, %d) returned an unexpected error: %v", tt.cost, tt.people, err)
			continue
		}
		if share != tt.wantShare || leftover != tt.wantLeftover {
			t.Errorf("Split(%d, %d)\n got: %d, %d\nwant: %d, %d", tt.cost, tt.people, share, leftover, tt.wantShare, tt.wantLeftover)
		}
	}
}
