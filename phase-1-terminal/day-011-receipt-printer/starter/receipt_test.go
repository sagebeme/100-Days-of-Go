package receiptprinter

import "testing"

func TestCommas(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{7, "7"},
		{999, "999"},
		{1000, "1,000"},
		{100000, "100,000"},
		{1250000, "1,250,000"},
		{-1000, "-1,000"},
		{-999, "-999"},
	}
	for _, tt := range tests {
		got := Commas(tt.n)
		if got != tt.want {
			t.Errorf("Commas(%d)\n got: %q\nwant: %q", tt.n, got, tt.want)
		}
	}
}

func TestShillings(t *testing.T) {
	tests := []struct {
		cents int
		want  string
	}{
		{125000, "Ksh1,250.00"},
		{5, "Ksh0.05"},
		{0, "Ksh0.00"},
		{99, "Ksh0.99"},
		{100, "Ksh1.00"},
		{123456789, "Ksh1,234,567.89"},
		{100000000, "Ksh1,000,000.00"},
		{-125050, "-Ksh1,250.50"},
	}
	for _, tt := range tests {
		got := Shillings(tt.cents)
		if got != tt.want {
			t.Errorf("Shillings(%d)\n got: %q\nwant: %q", tt.cents, got, tt.want)
		}
	}
}

func TestLine(t *testing.T) {
	tests := []struct {
		name  string
		qty   int
		price int
		want  string
	}{
		{"Unga 2kg", 2, 23000, "Unga 2kg           2      Ksh460.00"},
		{"Sukari ya Mumias 2kg", 1, 39000, "Sukari ya Mumias   1      Ksh390.00"},
		{"Bread", 120, 6500, "Bread            120    Ksh7,800.00"},
		{"Airtime", 1, 1000000, "Airtime            1   Ksh10,000.00"},
	}
	for _, tt := range tests {
		got := Line(tt.name, tt.qty, tt.price)
		if got != tt.want {
			t.Errorf("Line(%q, %d, %d)\n got: %q\nwant: %q", tt.name, tt.qty, tt.price, got, tt.want)
		}
	}
}

func TestReceipt(t *testing.T) {
	got, err := Receipt(
		[]string{"Unga 2kg", "Sukari 1kg", "Milk 500ml"},
		[]int{2, 1, 3},
		[]int{23000, 18500, 6000},
	)
	want := "Item        Qty  Total\n" +
		"Unga 2kg    2    Ksh460.00\n" +
		"Sukari 1kg  1    Ksh185.00\n" +
		"Milk 500ml  3    Ksh180.00\n" +
		"TOTAL            Ksh825.00\n"
	if err != nil || got != want {
		t.Errorf("Receipt with three items\n got: %q, %v\nwant: %q, <nil>\n\n got:\n%s\nwant:\n%s", got, err, want, got, want)
	}

	got, err = Receipt(nil, nil, nil)
	want = "Item   Qty  Total\nTOTAL       Ksh0.00\n"
	if err != nil || got != want {
		t.Errorf("Receipt with no items\n got: %q, %v\nwant: %q, <nil>", got, err, want)
	}
}

func TestReceiptRejects(t *testing.T) {
	tests := []struct {
		about  string
		items  []string
		qtys   []int
		prices []int
	}{
		{"lengths don't match", []string{"Unga", "Sukari"}, []int{1}, []int{100, 200}},
		{"zero quantity", []string{"Unga"}, []int{0}, []int{100}},
		{"negative price", []string{"Unga"}, []int{1}, []int{-100}},
	}
	for _, tt := range tests {
		if got, err := Receipt(tt.items, tt.qtys, tt.prices); err == nil {
			t.Errorf("Receipt with %s = %q with no error, want an error", tt.about, got)
		}
	}
}
