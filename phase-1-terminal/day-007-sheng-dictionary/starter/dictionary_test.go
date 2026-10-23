package shengdictionary

import (
	"maps"
	"slices"
	"testing"
)

func testDict() map[string]string {
	return map[string]string{
		"msee":  "guy",
		"keja":  "house",
		"fiti":  "fine",
		"doh":   "money",
		"mbogi": "crew",
		"poa":   "cool",
		"buda":  "dad",
		"form":  "plan",
		"empty": "",
	}
}

func TestLookup(t *testing.T) {
	dict := testDict()
	tests := []struct {
		word   string
		want   string
		wantOK bool
	}{
		{"keja", "house", true},
		{"msee", "guy", true},
		{"mbogi", "crew", true},
		{"rada", "", false},
		{"Keja", "", false},
		{"", "", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		got, ok := Lookup(dict, tt.word)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("Lookup(dict, %q)\n got: %q, %t\nwant: %q, %t", tt.word, got, ok, tt.want, tt.wantOK)
		}
	}

	if got, ok := Lookup(nil, "keja"); got != "" || ok {
		t.Errorf("Lookup(nil, %q)\n got: %q, %t\nwant: %q, %t", "keja", got, ok, "", false)
	}
}

func TestTranslateAll(t *testing.T) {
	dict := testDict()
	tests := []struct {
		words []string
		want  []string
	}{
		{[]string{"msee", "hii", "keja"}, []string{"guy", "hii", "house"}},
		{[]string{"buda", "ana", "doh"}, []string{"dad", "ana", "money"}},
		{[]string{"poa", "poa", "poa"}, []string{"cool", "cool", "cool"}},
		{[]string{"sasa", "rada"}, []string{"sasa", "rada"}},
		{[]string{"empty"}, []string{""}},
		{[]string{}, nil},
	}
	for _, tt := range tests {
		got := TranslateAll(dict, tt.words)
		if len(got) != len(tt.words) || !slices.Equal(got, tt.want) {
			t.Errorf("TranslateAll(dict, %q)\n got: %q\nwant: %q", tt.words, got, tt.want)
		}
	}
}

func TestWordCounts(t *testing.T) {
	tests := []struct {
		words []string
		want  map[string]int
	}{
		{[]string{"poa", "msee", "poa"}, map[string]int{"poa": 2, "msee": 1}},
		{[]string{"fiti"}, map[string]int{"fiti": 1}},
		{[]string{"a", "b", "a", "b", "a"}, map[string]int{"a": 3, "b": 2}},
		{[]string{"Poa", "poa"}, map[string]int{"Poa": 1, "poa": 1}},
		{[]string{}, map[string]int{}},
	}
	for _, tt := range tests {
		got := WordCounts(tt.words)
		if got == nil {
			t.Errorf("WordCounts(%q) returned nil, want an empty map you made with make", tt.words)
			continue
		}
		if !maps.Equal(got, tt.want) {
			t.Errorf("WordCounts(%q)\n got: %v\nwant: %v", tt.words, got, tt.want)
		}
	}
}

func TestUnknown(t *testing.T) {
	dict := testDict()
	tests := []struct {
		words []string
		want  []string
	}{
		{[]string{"msee", "rada", "rada"}, []string{"rada"}},
		{[]string{"sasa", "msee", "niaje", "sasa", "keja", "niaje"}, []string{"sasa", "niaje"}},
		{[]string{"keja", "poa", "doh"}, nil},
		{[]string{"empty", "zii"}, []string{"zii"}},
		{[]string{}, nil},
	}
	for _, tt := range tests {
		got := Unknown(dict, tt.words)
		if !slices.Equal(got, tt.want) {
			t.Errorf("Unknown(dict, %q)\n got: %q\nwant: %q", tt.words, got, tt.want)
		}
	}
}
