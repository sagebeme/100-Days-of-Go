package tabletests

import (
	"errors"
	"testing"
)

// assertCents fails the test if got isn't want. t.Helper() makes a failure point at
// the line that called assertCents, not at a line inside it.
func assertCents(t *testing.T, input string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("ParseAmount(%q)\n got: %d cents\nwant: %d cents", input, got, want)
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // cents
	}{
		{"plain shillings", "500", 50000},
		{"zero", "0", 0},
		{"one shilling", "1", 100},
		{"with Ksh", "Ksh500", 50000},
		{"with Ksh and a space", "Ksh 500", 50000},
		{"with KES", "KES 1250", 125000},
		{"prefix in any case", "kSH20", 2000},
		{"spaces around", "  Ksh 75  ", 7500},
		{"commas", "1,250", 125000},
		{"two commas", "12,345,678", 1234567800},
		{"commas and cents", "Ksh1,250.50", 125050},
		{"one decimal digit", "1250.5", 125050},
		{"cents only", "0.05", 5},
		{"largest allowed", "999,999,999.99", 99999999999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) returned an error: %v", tt.input, err)
			}
			assertCents(t, tt.input, got, tt.want)
		})
	}
}

func TestParseAmountErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"empty", "", ErrEmpty},
		{"only spaces", "   ", ErrEmpty},
		{"only the prefix", "Ksh", ErrEmpty},
		{"letters", "five hundred", ErrFormat},
		{"negative", "-500", ErrFormat},
		{"space inside", "1 250", ErrFormat},
		{"comma group too short", "12,50", ErrFormat},
		{"comma group too long", "1,2500", ErrFormat},
		{"first group too long", "1250,000", ErrFormat},
		{"trailing comma", "1,", ErrFormat},
		{"leading comma", ",500", ErrFormat},
		{"three decimal digits", "10.505", ErrFormat},
		{"dot with no cents", "10.", ErrFormat},
		{"dot with no shillings", ".50", ErrFormat},
		{"two dots", "1.2.3", ErrFormat},
		{"prefix twice", "KshKsh5", ErrFormat},
		{"prefix after the number", "500 Ksh", ErrFormat},
		{"ten digits", "1000000000", ErrTooLarge},
		{"ten digits with commas", "1,000,000,000.00", ErrTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if !errors.Is(err, tt.want) {
				t.Errorf("ParseAmount(%q)\n got: %d, %v\nwant: an error wrapping %q", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		cents int
		want  string
	}{
		{125050, "Ksh1,250.50"},
		{5, "Ksh0.05"},
		{0, "Ksh0.00"},
		{100, "Ksh1.00"},
		{99999, "Ksh999.99"},
		{100000, "Ksh1,000.00"},
		{1234567800, "Ksh12,345,678.00"},
		{-20000, "-Ksh200.00"},
		{-5, "-Ksh0.05"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatAmount(tt.cents); got != tt.want {
				t.Errorf("FormatAmount(%d)\n got: %q\nwant: %q", tt.cents, got, tt.want)
			}
		})
	}
}

// TestRoundTrip checks the two functions agree: formatting then parsing gives back the
// number you started with.
func TestRoundTrip(t *testing.T) {
	for _, cents := range []int{0, 1, 99, 100, 125050, 100000000, 99999999999} {
		text := FormatAmount(cents)
		got, err := ParseAmount(text)
		if err != nil {
			t.Errorf("ParseAmount(FormatAmount(%d)) = ParseAmount(%q) returned an error: %v", cents, text, err)
			continue
		}
		assertCents(t, text, got, cents)
	}
}

// FuzzRoundTrip is a fuzz test. A plain `go test` runs it once for each f.Add seed.
// `go test -fuzz=FuzzRoundTrip` keeps inventing new inputs until it finds one that
// fails or you press Ctrl+C.
func FuzzRoundTrip(f *testing.F) {
	for _, seed := range []int{0, 5, 125050, 99999999999} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, cents int) {
		if cents < 0 || cents > 99999999999 {
			t.Skip("ParseAmount only reads amounts from 0 to Ksh999,999,999.99")
		}
		text := FormatAmount(cents)
		got, err := ParseAmount(text)
		if err != nil || got != cents {
			t.Errorf("ParseAmount(FormatAmount(%d)) = ParseAmount(%q) = %d, %v", cents, text, got, err)
		}
	})
}
