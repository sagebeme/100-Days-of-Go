package pesa

import (
	"errors"
	"slices"
	"testing"
)

// These are the project's own tests, as they were when issue #42 was opened. They all
// pass. Add yours in a new file; don't change these.

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want Cents
	}{
		{"Ksh 1,234.50", 123450},
		{"KES 99", 9900},
		{"ksh99.00", 9900},
		{"1234", 123400},
		{"1,234,567.89", 123456789},
		{"  Ksh 20.05  ", 2005},
		{"-Ksh 20.00", -2000},
		{"Ksh -20.00", -2000},
		{"0", 0},
		{"0.00", 0},
		{"9999999999999.99", 999999999999999},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("Parse(%q) = %d, %v; want %d, nil", tt.in, got, err, tt.want)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{
		"", "Ksh", "abc", "12a", "1.234", "1.", ".50", "1,23", "12,3456", ",123", "1,,234",
		"--5", "-Ksh -5", "Ksh 1 000", "USD 5", "1.2.3", "10000000000000",
	} {
		if got, err := Parse(in); !errors.Is(err, ErrSyntax) {
			t.Errorf("Parse(%q) = %d, %v; want ErrSyntax", in, got, err)
		}
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		in   Cents
		want string
	}{
		{123450, "Ksh 1,234.50"},
		{9900, "Ksh 99.00"},
		{5, "Ksh 0.05"},
		{0, "Ksh 0.00"},
		{-2000, "-Ksh 20.00"},
		{100000000, "Ksh 1,000,000.00"},
	}
	for _, tt := range tests {
		if got := Format(tt.in); got != tt.want {
			t.Errorf("Format(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		c    Cents
		n    int
		want []Cents
	}{
		{100000, 4, []Cents{25000, 25000, 25000, 25000}},
		{100, 3, []Cents{34, 33, 33}},
		{-100, 3, []Cents{-34, -33, -33}},
		{2, 5, []Cents{1, 1, 0, 0, 0}},
	}
	for _, tt := range tests {
		got := Split(tt.c, tt.n)
		if !slices.Equal(got, tt.want) || Sum(got...) != tt.c {
			t.Errorf("Split(%d, %d) = %v, want %v", tt.c, tt.n, got, tt.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, c := range []Cents{0, 1, 10, 99, 100, 123450, -2005, 999999999999999} {
		got, err := Parse(Format(c))
		if err != nil || got != c {
			t.Errorf("Parse(Format(%d)) = %d, %v", c, got, err)
		}
	}
}
