package mpesasms

import "testing"

func TestBuildConfirmation(t *testing.T) {
	tests := []struct {
		code   string
		amount int
		name   string
		want   string
	}{
		{"QFT1ABC2DE", 500, "Amina Wanjiru", "QFT1ABC2DE Confirmed. Ksh500.00 sent to Amina Wanjiru."},
		{"RKD4XY7Z1P", 1250, "Kip", "RKD4XY7Z1P Confirmed. Ksh1250.00 sent to Kip."},
		{"SAB0000001", 1, "Mama Mboga", "SAB0000001 Confirmed. Ksh1.00 sent to Mama Mboga."},
		{"TZZ9QWE8RT", 0, "Otieno", "TZZ9QWE8RT Confirmed. Ksh0.00 sent to Otieno."},
	}
	for _, tt := range tests {
		got := BuildConfirmation(tt.code, tt.amount, tt.name)
		if got != tt.want {
			t.Errorf("BuildConfirmation(%q, %d, %q)\n got: %q\nwant: %q", tt.code, tt.amount, tt.name, got, tt.want)
		}
	}
}
