package packageitup

import "testing"

var prices = map[string]int{
	"unga":   230,
	"sukari": 185,
	"maziwa": 60,
	"mkate":  65,
	"sukuma": 20,
	"gas":    3200,
}

func TestTotal(t *testing.T) {
	tests := []struct {
		basket map[string]int
		want   int
	}{
		{map[string]int{"unga": 2}, 460},
		{map[string]int{"unga": 2, "sukari": 1}, 645},
		{map[string]int{"maziwa": 3, "mkate": 1, "sukuma": 5}, 345},
		{map[string]int{"gas": 1000}, 3200000},
		{map[string]int{}, 0},
	}
	for _, tt := range tests {
		got, err := Total(tt.basket, prices)
		if err != nil || got != tt.want {
			t.Errorf("Total(%v, prices)\n got: %d, %v\nwant: %d, <nil>", tt.basket, got, err, tt.want)
		}
	}
}

func TestTotalRejects(t *testing.T) {
	bad := []map[string]int{
		{"nyama": 1},
		{"unga": 1, "chapati": 2},
		{"unga": 0},
		{"sukari": -2},
	}
	for _, basket := range bad {
		if got, err := Total(basket, prices); err == nil {
			t.Errorf("Total(%v, prices) = %d with no error, want an error", basket, got)
		}
	}
}

func TestSummary(t *testing.T) {
	tests := []struct {
		basket map[string]int
		want   string
	}{
		{map[string]int{"unga": 2, "sukari": 1}, "3 items, Ksh645"},
		{map[string]int{"sukuma": 1}, "1 item, Ksh20"},
		{map[string]int{"unga": 10, "sukari": 2}, "12 items, Ksh2,670"},
		{map[string]int{"gas": 1000}, "1000 items, Ksh3,200,000"},
		{map[string]int{"gas": 1, "maziwa": 1}, "2 items, Ksh3,260"},
		{map[string]int{}, "0 items, Ksh0"},
	}
	for _, tt := range tests {
		got, err := Summary(tt.basket, prices)
		if err != nil || got != tt.want {
			t.Errorf("Summary(%v, prices)\n got: %q, %v\nwant: %q, <nil>", tt.basket, got, err, tt.want)
		}
	}

	if got, err := Summary(map[string]int{"nyama": 1}, prices); err == nil || got != "" {
		t.Errorf("Summary with an unknown item = %q, %v; want \"\" and an error", got, err)
	}
}
