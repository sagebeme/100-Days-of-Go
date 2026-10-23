package examgrader

import (
	"slices"
	"testing"
)

func TestMean(t *testing.T) {
	tests := []struct {
		scores []int
		want   float64
	}{
		{[]int{70, 80, 90}, 80},
		{[]int{50, 51}, 50.5},
		{[]int{100}, 100},
		{[]int{0, 0, 0}, 0},
		{[]int{1, 2, 2}, 5.0 / 3},
		{nil, 0},
	}
	for _, tt := range tests {
		got := Mean(tt.scores)
		if got != tt.want {
			t.Errorf("Mean(%v) = %v, want %v", tt.scores, got, tt.want)
		}
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		scores []int
		want   float64
	}{
		{[]int{90, 40, 70}, 70},
		{[]int{10, 40, 30, 20}, 25},
		{[]int{55, 56}, 55.5},
		{[]int{77}, 77},
		{[]int{80, 80, 20, 80}, 80},
		{nil, 0},
	}
	for _, tt := range tests {
		got := Median(tt.scores)
		if got != tt.want {
			t.Errorf("Median(%v) = %v, want %v", tt.scores, got, tt.want)
		}
	}
}

func TestMedianLeavesInputAlone(t *testing.T) {
	scores := []int{90, 40, 70}
	Median(scores)
	if !slices.Equal(scores, []int{90, 40, 70}) {
		t.Errorf("Median changed its input to %v; sort a copy instead", scores)
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "A"}, {80, "A"}, {79, "B"}, {65, "B"}, {64, "C"},
		{50, "C"}, {49, "D"}, {40, "D"}, {39, "E"}, {0, "E"},
	}
	for _, tt := range tests {
		got, err := Grade(tt.score)
		if err != nil || got != tt.want {
			t.Errorf("Grade(%d)\n got: %q, %v\nwant: %q, <nil>", tt.score, got, err, tt.want)
		}
	}
	for _, score := range []int{-1, 101, 250} {
		if got, err := Grade(score); err == nil {
			t.Errorf("Grade(%d) = %q with no error, want an error", score, got)
		}
	}
}

func TestRank(t *testing.T) {
	tests := []struct {
		scores map[string]int
		want   []string
	}{
		{
			map[string]int{"Amina": 70, "Baraka": 85, "Chebet": 70},
			[]string{"Baraka", "Amina", "Chebet"},
		},
		{
			map[string]int{"Wanjiku": 55, "Otieno": 91, "Kip": 38, "Achieng": 91, "Mwangi": 60},
			[]string{"Achieng", "Otieno", "Mwangi", "Wanjiku", "Kip"},
		},
		{
			map[string]int{"Zawadi": 0},
			[]string{"Zawadi"},
		},
		{
			map[string]int{},
			[]string{},
		},
	}
	for _, tt := range tests {
		got := Rank(tt.scores)
		if !slices.Equal(got, tt.want) {
			t.Errorf("Rank(%v)\n got: %#v\nwant: %#v", tt.scores, got, tt.want)
		}
	}
}
