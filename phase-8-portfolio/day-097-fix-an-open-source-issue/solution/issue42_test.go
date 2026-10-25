package pesa

import "testing"

// TestParseOneFractionDigit reproduces issue #42: written before the fix, it failed
// with Parse("Ksh 1,200.5") = 120005 (Ksh 1,200.05).
func TestParseOneFractionDigit(t *testing.T) {
	tests := []struct {
		in   string
		want Cents
	}{
		{"Ksh 1,200.5", 120050}, // the exact input from the issue
		{"0.5", 50},
		{"-0.5", -50},
		{"1.05", 105}, // two digits must not change
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("Parse(%q) = %d, %v; want %d", tt.in, got, err, tt.want)
		}
	}
}
