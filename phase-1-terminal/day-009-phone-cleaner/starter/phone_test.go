package phonecleaner

import "testing"

func TestCleanAccepts(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"0712 345 678", "254712345678"},
		{"+254712345678", "254712345678"},
		{"254712345678", "254712345678"},
		{"0712-345-678", "254712345678"},
		{"  0712345678  ", "254712345678"},
		{"712345678", "254712345678"},
		{"0110 123 456", "254110123456"},
		{"+254 110 123456", "254110123456"},
		{"0799999999", "254799999999"},
	}
	for _, tt := range tests {
		got, err := Clean(tt.raw)
		if err != nil {
			t.Errorf("Clean(%q) returned error %q, want %q", tt.raw, err, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("Clean(%q)\n got: %q\nwant: %q", tt.raw, got, tt.want)
		}
	}
}

func TestCleanRejects(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"071234567",     // only 9 digits, and it starts with 0
		"07123456789",   // 11 digits
		"0712a45678",    // a letter
		"(0712) 345678", // brackets
		"+0712345678",   // + must be followed by 254
		"0712+345678",   // + in the middle
		"0201234567",    // a Nairobi landline, not a mobile
		"255712345678",  // a Tanzanian number
		"2547123456789", // too long
		"+254 812 345 678",
	}
	for _, raw := range bad {
		got, err := Clean(raw)
		if err == nil {
			t.Errorf("Clean(%q) = %q with no error, want an error", raw, got)
			continue
		}
		if got != "" {
			t.Errorf("Clean(%q) returned %q with its error, want \"\"", raw, got)
		}
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"+254712345678", "0712***678"},
		{"0110123456", "0110***456"},
		{"722 000 111", "0722***111"},
	}
	for _, tt := range tests {
		got, err := Mask(tt.raw)
		if err != nil || got != tt.want {
			t.Errorf("Mask(%q)\n got: %q, %v\nwant: %q, <nil>", tt.raw, got, err, tt.want)
		}
	}

	if got, err := Mask("0712"); err == nil {
		t.Errorf("Mask(%q) = %q with no error, want an error", "0712", got)
	}
}

func TestSameNumber(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"0712 345 678", "+254712345678", true},
		{"712345678", "254712345678", true},
		{"0712345678", "0712345679", false},
		{"abc", "abc", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got := SameNumber(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("SameNumber(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
