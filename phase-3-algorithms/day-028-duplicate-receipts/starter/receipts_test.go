package duplicatereceipts

import (
	"maps"
	"strconv"
	"testing"
)

func TestFirstDuplicate(t *testing.T) {
	tests := []struct {
		name   string
		codes  []string
		want   string
		wantOK bool
	}{
		{"empty", nil, "", false},
		{"one receipt", []string{"QFT1"}, "", false},
		{"no repeats", []string{"QFT1", "RKD4", "SAB0"}, "", false},
		{"repeat right away", []string{"QFT1", "QFT1"}, "QFT1", true},
		{"second sighting decides", []string{"QFT1", "RKD4", "SAB0", "RKD4", "QFT1"}, "RKD4", true},
		{"case matters", []string{"qft1", "QFT1"}, "", false},
		{"three of a kind", []string{"A", "B", "A", "A"}, "A", true},
		{"empty code counts too", []string{"", "X", ""}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FirstDuplicate(tt.codes)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("FirstDuplicate(%q) = %q, %v, want %q, %v", tt.codes, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestFirstDuplicateBig(t *testing.T) {
	// 200 000 codes, all different except the last one. An O(n²) scan does 2 × 10^10 checks.
	codes := make([]string, 200_000)
	for i := range codes {
		codes[i] = "R" + strconv.Itoa(i)
	}
	codes[len(codes)-1] = "R123"
	got, ok := FirstDuplicate(codes)
	if got != "R123" || !ok {
		t.Errorf("FirstDuplicate(200000 codes) = %q, %v, want %q, true", got, ok, "R123")
	}
}

func TestCountFrequencies(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  map[string]int
	}{
		{"empty", []string{}, map[string]int{}},
		{"one", []string{"unga"}, map[string]int{"unga": 1}},
		{"repeats", []string{"unga", "sukari", "unga", "unga"}, map[string]int{"unga": 3, "sukari": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountFrequencies(tt.items)
			if got == nil {
				t.Fatalf("CountFrequencies(%q) returned a nil map", tt.items)
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("CountFrequencies(%q)\n got: %v\nwant: %v", tt.items, got, tt.want)
			}
		})
	}
}

func TestTopSeller(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		want      string
		wantCount int
	}{
		{"empty", nil, "", 0},
		{"one sale", []string{"samosa"}, "samosa", 1},
		{"clear winner", []string{"chai", "mandazi", "chai"}, "chai", 2},
		{"tie goes to the alphabet", []string{"mandazi", "chai", "chai", "mandazi", "samosa"}, "chai", 2},
		{"all different", []string{"z", "y", "x"}, "x", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := TopSeller(tt.items)
			if got != tt.want || n != tt.wantCount {
				t.Errorf("TopSeller(%q) = %q, %d, want %q, %d", tt.items, got, n, tt.want, tt.wantCount)
			}
		})
	}
}

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name    string
		amounts []int
		target  int
		i, j    int
		ok      bool
	}{
		{"empty", nil, 100, 0, 0, false},
		{"one amount can't pair with itself", []int{50}, 100, 0, 0, false},
		{"first two", []int{300, 700, 150, 850}, 1000, 0, 1, true},
		{"smallest j wins", []int{150, 300, 850, 700}, 1000, 0, 2, true},
		{"duplicates make a pair", []int{500, 20, 500}, 1000, 0, 2, true},
		{"earliest i for that j", []int{400, 400, 600}, 1000, 0, 2, true},
		{"negatives (refunds)", []int{-200, 50, 1200}, 1000, 0, 2, true},
		{"zero target", []int{5, 0, -5}, 0, 0, 2, true},
		{"no pair", []int{1, 2, 3}, 100, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, j, ok := TwoSum(tt.amounts, tt.target)
			if i != tt.i || j != tt.j || ok != tt.ok {
				t.Errorf("TwoSum(%v, %d) = %d, %d, %v, want %d, %d, %v", tt.amounts, tt.target, i, j, ok, tt.i, tt.j, tt.ok)
			}
		})
	}
}

func TestTwoSumBig(t *testing.T) {
	a := make([]int, 200_000)
	for i := range a {
		a[i] = 2 * i // all even, so an odd target never matches
	}
	if _, _, ok := TwoSum(a, 1); ok {
		t.Errorf("TwoSum(200000 even numbers, 1) found a pair, want none")
	}
	a[len(a)-1] = 1 - a[7]
	if i, j, ok := TwoSum(a, 1); i != 7 || j != len(a)-1 || !ok {
		t.Errorf("TwoSum(big, 1) = %d, %d, %v, want 7, %d, true", i, j, ok, len(a)-1)
	}
}
