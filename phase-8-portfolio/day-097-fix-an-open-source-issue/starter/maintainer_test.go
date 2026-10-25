package pesa

import "testing"

// The maintainers added these tests when they merged the fix for issue #42. They are
// the course's check that your fix is right. Write your own reproduction test first,
// in a file of your own, before you read this one.

func TestIssue42OneDecimalDigit(t *testing.T) {
	tests := []struct {
		in   string
		want Cents
	}{
		{"0.5", 50},
		{"Ksh 1,200.5", 120050},
		{"KES 99.9", 9990},
		{"-0.5", -50},
		{"Ksh -3.1", -310},
		{"1.05", 105}, // two digits still mean cents
		{"1.50", 150},
		{"1.0", 100},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("Parse(%q) = %d (%s), %v; want %d (%s)", tt.in, got, Format(got), err, tt.want, Format(tt.want))
		}
	}
}

func TestIssue42FormatRoundTrip(t *testing.T) {
	// A one-digit amount, read and written back, must come out as the same money.
	for in, want := range map[string]string{"1200.5": "Ksh 1,200.50", "0.1": "Ksh 0.10", "-7.7": "-Ksh 7.70"} {
		c, err := Parse(in)
		if err != nil || Format(c) != want {
			t.Errorf("Format(Parse(%q)) = %q, %v; want %q", in, Format(c), err, want)
		}
	}
}
