package statement

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestSummarise(t *testing.T) {
	in := "receipt,date,counterparty,amount\n" +
		"SJK2X1ABCD,2026-10-01 08:15,MAMA MBOGA KAWANGWARE,-350.00\n" +
		"SJK2X1ABCE,2026-10-01 09:00,Otieno Wekesa,1500.00\r\n" +
		"\n" +
		"SJK2X1ABCF,2026-10-02 18:30,  mama mboga kawangware ,-120.50\n" +
		"SJK2X1ABCG,2026-10-03 07:45,KPLC PREPAID,-1000.00\n" +
		"SJK2X1ABCH,2026-10-03 12:00,OTIENO WEKESA,-200.00\n" +
		"SJK2X1ABCI,2026-10-04 10:10,CHAMA YA MAMA,0.00\n"
	want := []Total{
		{Name: "KPLC PREPAID", Count: 1, OutCents: 100000},
		{Name: "MAMA MBOGA KAWANGWARE", Count: 2, OutCents: 47050},
		{Name: "OTIENO WEKESA", Count: 2, InCents: 150000, OutCents: 20000},
		{Name: "CHAMA YA MAMA", Count: 1},
	}
	got, err := Summarise(in)
	if err != nil {
		t.Fatalf("Summarise: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Summarise\n got: %+v\nwant: %+v", got, want)
	}
}

func TestSortTies(t *testing.T) {
	in := "AAAAAAAAA1,d,ZEBRA,-100.00\n" +
		"AAAAAAAAA2,d,APPLE,-100.00\n" +
		"AAAAAAAAA3,d,MANGO,-100.00\n" +
		"AAAAAAAAA4,d,MANGO,50.00\n" +
		"AAAAAAAAA5,d,BANANA,999999999999.99"
	got, err := Summarise(in)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tt := range got {
		names = append(names, tt.Name)
	}
	if want := []string{"MANGO", "APPLE", "ZEBRA", "BANANA"}; !slices.Equal(names, want) {
		t.Errorf("order = %v, want %v (out desc, then in desc, then name)", names, want)
	}
	if got[3].InCents != 99999999999999 {
		t.Errorf("BANANA InCents = %d, want 99999999999999 exactly", got[3].InCents)
	}
}

func TestEmpty(t *testing.T) {
	for _, in := range []string{"", "\n\n", "receipt,date,counterparty,amount\n"} {
		got, err := Summarise(in)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("Summarise(%q) = %v, %v; want an empty, non-nil slice", in, got, err)
		}
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"short receipt", "SJK2X1ABC,2026-10-01,DUKA,-1.00"},
		{"lower-case receipt", "sjk2x1abcd,2026-10-01,DUKA,-1.00"},
		{"no date", "SJK2X1ABCD,,DUKA,-1.00"},
		{"blank counterparty", "SJK2X1ABCD,2026-10-01,   ,-1.00"},
		{"no cents", "SJK2X1ABCD,2026-10-01,DUKA,-350"},
		{"one cent digit", "SJK2X1ABCD,2026-10-01,DUKA,-350.5"},
		{"plus sign", "SJK2X1ABCD,2026-10-01,DUKA,+350.00"},
		{"13 digits", "SJK2X1ABCD,2026-10-01,DUKA,1234567890123.00"},
		{"too many fields", "SJK2X1ABCD,2026-10-01,DUKA,-1.00,extra"},
		{"too few fields", "SJK2X1ABCD,2026-10-01,DUKA"},
		{"letters in amount", "SJK2X1ABCD,2026-10-01,DUKA,-1a.00"},
	}
	for _, tt := range tests {
		in := "receipt,date,counterparty,amount\nSJK2X1AAAA,2026-10-01,DUKA,-5.00\n" + tt.line + "\n"
		_, err := Summarise(in)
		var le *LineError
		if !errors.As(err, &le) || le.Line != 3 {
			t.Errorf("%s: Summarise = %v, want a *LineError for line 3", tt.name, err)
		}
	}
}

// statement builds a statement of n lines over k counterparties, and the Totals it
// should give, worked out as it goes.
func statement(n, k int) (string, []Total) {
	var b strings.Builder
	b.WriteString("receipt,date,counterparty,amount\n")
	byName := map[string]*Total{}
	for i := range n {
		name := fmt.Sprintf("SHOP %03d NAIROBI", (i*7)%k)
		cents := int64((i*7919)%2000000) - 1500000
		fmt.Fprintf(&b, "S%09d,2026-10-%02d 08:%02d,%s,", i, i%28+1, i%60, name)
		if cents < 0 {
			fmt.Fprintf(&b, "-%d.%02d\n", -cents/100, -cents%100)
		} else {
			fmt.Fprintf(&b, "%d.%02d\n", cents/100, cents%100)
		}
		t := byName[name]
		if t == nil {
			t = &Total{Name: name}
			byName[name] = t
		}
		t.Count++
		if cents < 0 {
			t.OutCents -= cents
		} else {
			t.InCents += cents
		}
	}
	var want []Total
	for _, t := range byName {
		want = append(want, *t)
	}
	slices.SortFunc(want, func(a, b Total) int {
		switch {
		case a.OutCents != b.OutCents:
			if a.OutCents > b.OutCents {
				return -1
			}
			return 1
		case a.InCents != b.InCents:
			if a.InCents > b.InCents {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return b.String(), want
}

func TestLargeStatement(t *testing.T) {
	in, want := statement(2000, 200)
	got, err := Summarise(in)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Summarise of 2000 lines: got %d totals, want %d; first got %+v, first want %+v", len(got), len(want), got[0], want[0])
	}
}

// The performance guards. They count allocations, which is the same on every machine
// and every run, so they never flake the way timing does. Speed itself you measure
// with the benchmark below.

func TestAllocationsAreBounded(t *testing.T) {
	in, _ := statement(2000, 50)
	allocs := testing.AllocsPerRun(1, func() {
		if _, err := Summarise(in); err != nil {
			t.Fatal(err)
		}
	})
	// A one-pass version needs a few allocations per counterparty (the map and the
	// slice growing), and none per line.
	if allocs > 200 {
		t.Errorf("Summarise of 2000 lines over 50 counterparties made %.0f allocations, want at most 200", allocs)
	}
}

func TestAllocationsDontGrowWithLines(t *testing.T) {
	small, _ := statement(500, 50)
	big, _ := statement(2000, 50)
	count := func(in string) float64 {
		return testing.AllocsPerRun(1, func() { Summarise(in) })
	}
	a, b := count(small), count(big)
	// Four times the lines with the same 50 names should cost about the same: the work
	// per line must not allocate.
	if b > a+20 {
		t.Errorf("allocations went from %.0f for 500 lines to %.0f for 2000; want no growth per line", a, b)
	}
}

func BenchmarkSummarise(b *testing.B) {
	in, _ := statement(5000, 200)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		if _, err := Summarise(in); err != nil {
			b.Fatal(err)
		}
	}
}
