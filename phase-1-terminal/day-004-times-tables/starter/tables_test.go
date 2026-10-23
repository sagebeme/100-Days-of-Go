package timestables

import "testing"

func TestTable(t *testing.T) {
	tests := []struct {
		n, upTo int
		want    string
	}{
		{3, 3, "3 x 1 = 3\n3 x 2 = 6\n3 x 3 = 9\n"},
		{7, 1, "7 x 1 = 7\n"},
		{12, 2, "12 x 1 = 12\n12 x 2 = 24\n"},
		{0, 2, "0 x 1 = 0\n0 x 2 = 0\n"},
		{5, 0, ""},
		{5, -3, ""},
	}
	for _, tt := range tests {
		got := Table(tt.n, tt.upTo)
		if got != tt.want {
			t.Errorf("Table(%d, %d)\n got: %q\nwant: %q", tt.n, tt.upTo, got, tt.want)
		}
	}

	// A full table, checked line by line.
	full := Table(9, 12)
	if len(full) == 0 || full[len(full)-1] != '\n' {
		t.Errorf("Table(9, 12) should end with a newline, got: %q", full)
	}
	wantEnd := "9 x 12 = 108\n"
	if len(full) < len(wantEnd) || full[len(full)-len(wantEnd):] != wantEnd {
		t.Errorf("Table(9, 12) should end with %q, got: %q", wantEnd, full)
	}
}

func TestSumTable(t *testing.T) {
	tests := []struct {
		n, upTo int
		want    int
	}{
		{7, 12, 546},
		{3, 3, 18},
		{1, 100, 5050},
		{9, 1, 9},
		{0, 12, 0},
		{5, 0, 0},
		{-2, 3, -12},
	}
	for _, tt := range tests {
		got := SumTable(tt.n, tt.upTo)
		if got != tt.want {
			t.Errorf("SumTable(%d, %d)\n got: %d\nwant: %d", tt.n, tt.upTo, got, tt.want)
		}
	}
}

func TestCountDivisible(t *testing.T) {
	tests := []struct {
		from, to, d int
		want        int
	}{
		{1, 100, 3, 33},
		{1, 10, 1, 10},
		{0, 10, 5, 3},
		{7, 7, 7, 1},
		{8, 8, 7, 0},
		{1, 6, 7, 0},
		{10, 1, 2, 0},
		{-10, 10, 5, 5},
	}
	for _, tt := range tests {
		got := CountDivisible(tt.from, tt.to, tt.d)
		if got != tt.want {
			t.Errorf("CountDivisible(%d, %d, %d)\n got: %d\nwant: %d", tt.from, tt.to, tt.d, got, tt.want)
		}
	}
}
