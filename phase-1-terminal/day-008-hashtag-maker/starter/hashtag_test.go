package hashtagmaker

import "testing"

func TestHashtag(t *testing.T) {
	tests := []struct {
		phrase string
		want   string
	}{
		{"nairobi tech week", "#NairobiTechWeek"},
		{"harambee stars", "#HarambeeStars"},
		{"  MASHUJAA day  2026! ", "#MashujaaDay2026"},
		{"2026 elections", "#2026Elections"},
		{"sema ng'ware", "#SemaNgware"},
		{"café ya mama", "#CaféYaMama"},
		{"éLIMU bora", "#ÉlimuBora"},
		{"kenya", "#Kenya"},
		{"", ""},
		{"!!! ???", ""},
	}
	for _, tt := range tests {
		got := Hashtag(tt.phrase)
		if got != tt.want {
			t.Errorf("Hashtag(%q)\n got: %q\nwant: %q", tt.phrase, got, tt.want)
		}
	}
}

func TestInitials(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"amina wanjiru", "AW"},
		{"Ngũgĩ wa Thiong'o", "NWT"},
		{"  kip  ", "K"},
		{"émile ñambura", "ÉÑ"},
		{"", ""},
	}
	for _, tt := range tests {
		got := Initials(tt.name)
		if got != tt.want {
			t.Errorf("Initials(%q)\n got: %q\nwant: %q", tt.name, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"Habari za asubuhi", 6, "Habari..."},
		{"Jambo", 5, "Jambo"},
		{"Jambo", 10, "Jambo"},
		{"café latte", 4, "café..."},
		{"Ngũgĩ", 3, "Ngũ..."},
		{"Sasa", 0, "..."},
		{"", 3, ""},
	}
	for _, tt := range tests {
		got := Truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d)\n got: %q\nwant: %q", tt.s, tt.max, got, tt.want)
		}
	}
}
