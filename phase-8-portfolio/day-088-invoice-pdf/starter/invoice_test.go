package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func sample() Invoice {
	return Invoice{
		Number: "INV-2026-0042",
		Date:   time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC),
		Seller: Party{Name: "Mama Njeri Hardware Ltd", KRAPIN: "P051234567Q", Address: "Moi Avenue, Nairobi", Phone: "0712 345 678"},
		Buyer:  Party{Name: "Kamau (Builders) & Sons", KRAPIN: "A123456789B"},
		Lines: []Line{
			{Description: "Cement, 50 kg bag", Qty: 10, UnitCents: 75000},
			{Description: "Iron sheets, gauge 30", Qty: 4, UnitCents: 89950},
			{Description: "Nails 3\" (1 kg)", Qty: 3, UnitCents: 18000},
		},
	}
}

func TestComputeTotals(t *testing.T) {
	tests := []struct {
		name  string
		lines []Line
		want  Totals
	}{
		{"one line", []Line{{"x", 1, 100000}}, Totals{100000, 16000, 116000}},
		{"sample", sample().Lines, Totals{1163800, 186208, 1350008}},
		{"0.48 cents rounds down", []Line{{"x", 1, 3}}, Totals{3, 0, 3}},
		{"0.64 cents rounds up", []Line{{"x", 1, 4}}, Totals{4, 1, 5}},
		{"749.92 cents rounds up", []Line{{"x", 1, 4687}}, Totals{4687, 750, 5437}},
		{"VAT on the subtotal, not per line", []Line{{"x", 1, 3}, {"y", 1, 3}}, Totals{6, 1, 7}},
		{"free line", []Line{{"x", 5, 0}}, Totals{0, 0, 0}},
	}
	for _, tt := range tests {
		if got := ComputeTotals(tt.lines); got != tt.want {
			t.Errorf("%s: ComputeTotals = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestVATRoundsToNearestCent(t *testing.T) {
	for sub := int64(0); sub < 5000; sub++ {
		got := ComputeTotals([]Line{{"x", 1, sub}}).VAT
		want := int64(math.Round(float64(sub) * 0.16))
		if got != want {
			t.Fatalf("VAT on %d cents = %d, want %d", sub, got, want)
		}
	}
}

func TestValidKRAPIN(t *testing.T) {
	tests := []struct {
		pin  string
		want bool
	}{
		{"P051234567Q", true},
		{"A123456789B", true},
		{"", false},
		{"P05123456Q", false},
		{"P0512345678Q", false},
		{"B051234567Q", false},
		{"p051234567q", false},
		{"P05123A567Q", false},
		{"P0512345678", false},
	}
	for _, tt := range tests {
		if got := ValidKRAPIN(tt.pin); got != tt.want {
			t.Errorf("ValidKRAPIN(%q) = %v, want %v", tt.pin, got, tt.want)
		}
	}
}

func TestFormatKsh(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{0, "Ksh 0.00"},
		{5, "Ksh 0.05"},
		{116000, "Ksh 1,160.00"},
		{99999, "Ksh 999.99"},
		{100000, "Ksh 1,000.00"},
		{1350008, "Ksh 13,500.08"},
		{123456789012, "Ksh 1,234,567,890.12"},
		{-5, "-Ksh 0.05"},
	}
	for _, tt := range tests {
		if got := FormatKsh(tt.cents); got != tt.want {
			t.Errorf("FormatKsh(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func TestEscape(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Cement", "Cement"},
		{"Kamau (Builders)", `Kamau \(Builders\)`},
		{`C:\path`, `C:\\path`},
		{"Café", "Caf?"},
		{"line\nbreak", "line?break"},
	}
	for _, tt := range tests {
		if got := Escape(tt.in); got != tt.want {
			t.Errorf("Escape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTextWidth(t *testing.T) {
	if got := TextWidth("1,160.00", 10); got != 38.92 {
		t.Errorf("TextWidth(1,160.00, 10) = %v, want 38.92", got)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Invoice)
		want   error
	}{
		{"valid", func(*Invoice) {}, nil},
		{"no buyer PIN is fine", func(i *Invoice) { i.Buyer.KRAPIN = "" }, nil},
		{"no number", func(i *Invoice) { i.Number = " " }, ErrNoNumber},
		{"no date", func(i *Invoice) { i.Date = time.Time{} }, ErrNoDate},
		{"no seller name", func(i *Invoice) { i.Seller.Name = "" }, ErrSellerName},
		{"no seller PIN", func(i *Invoice) { i.Seller.KRAPIN = "" }, ErrSellerPIN},
		{"bad buyer PIN", func(i *Invoice) { i.Buyer.KRAPIN = "12345" }, ErrBuyerPIN},
		{"no lines", func(i *Invoice) { i.Lines = nil }, ErrNoLines},
		{"too many lines", func(i *Invoice) { i.Lines = make([]Line, MaxLines+1) }, ErrTooManyLines},
		{"zero qty", func(i *Invoice) { i.Lines[0].Qty = 0 }, ErrBadLine},
		{"negative price", func(i *Invoice) { i.Lines[0].UnitCents = -1 }, ErrBadLine},
		{"blank description", func(i *Invoice) { i.Lines[0].Description = "  " }, ErrBadLine},
		{"huge amount", func(i *Invoice) { i.Lines[0].Qty, i.Lines[0].UnitCents = 1<<31-1, 1<<40 }, ErrTooLarge},
	}
	for _, tt := range tests {
		inv := sample()
		tt.change(&inv)
		if err := inv.Validate(); !errors.Is(err, tt.want) {
			t.Errorf("%s: Validate() = %v, want %v", tt.name, err, tt.want)
		}
	}
}

func TestRenderRefusesInvalid(t *testing.T) {
	inv := sample()
	inv.Seller.KRAPIN = "nope"
	if b, err := Render(inv); !errors.Is(err, ErrSellerPIN) || b != nil {
		t.Errorf("Render(invalid) = %d bytes, %v, want nil, ErrSellerPIN", len(b), err)
	}
}

// TestRenderStructure checks the file the way a PDF reader does: from the end.
func TestRenderStructure(t *testing.T) {
	pdf, err := Render(sample())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.")) {
		t.Fatalf("file starts with %q, want %%PDF-1.", pdf[:min(8, len(pdf))])
	}
	if !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Fatalf("file doesn't end with %%%%EOF")
	}

	// startxref gives the offset of the xref table.
	m := regexp.MustCompile(`startxref\n(\d+)\n%%EOF\n$`).FindSubmatch(pdf)
	if m == nil {
		t.Fatalf("no startxref line before %%%%EOF")
	}
	xref, _ := strconv.Atoi(string(m[1]))
	if xref <= 0 || xref >= len(pdf) || !bytes.HasPrefix(pdf[xref:], []byte("xref\n")) {
		t.Fatalf("startxref = %d, but the bytes there aren't the xref table", xref)
	}

	// The xref table: "0 N", then N entries of exactly 20 bytes each.
	rest := pdf[xref+len("xref\n"):]
	header, rest, _ := bytes.Cut(rest, []byte("\n"))
	var first, count int
	if _, err := fmt.Sscanf(string(header), "%d %d", &first, &count); err != nil || first != 0 {
		t.Fatalf("xref subsection header %q, want \"0 N\"", header)
	}
	if len(rest) < 20*count {
		t.Fatalf("xref table is cut short")
	}
	if got := string(rest[:20]); got != "0000000000 65535 f \n" {
		t.Errorf("xref entry 0 = %q, want the free-list head", got)
	}
	for i := 1; i < count; i++ {
		entry := string(rest[20*i : 20*i+20])
		if !regexp.MustCompile(`^\d{10} 00000 n \n$`).MatchString(entry) {
			t.Fatalf("xref entry %d = %q, want 20 bytes like \"0000000015 00000 n \\n\"", i, entry)
		}
		off, _ := strconv.Atoi(entry[:10])
		want := fmt.Sprintf("%d 0 obj", i)
		if off >= len(pdf) || !bytes.HasPrefix(pdf[off:], []byte(want)) {
			t.Errorf("xref says object %d is at %d, but the bytes there are %q", i, off, pdf[off:min(off+12, len(pdf))])
		}
	}

	trailer := rest[20*count:]
	if !bytes.HasPrefix(trailer, []byte("trailer")) {
		t.Fatalf("no trailer right after the xref table")
	}
	if !bytes.Contains(trailer, []byte(fmt.Sprintf("/Size %d", count))) {
		t.Errorf("trailer %q, want /Size %d", trailer, count)
	}
	if !regexp.MustCompile(`/Root (\d+) 0 R`).Match(trailer) {
		t.Errorf("trailer has no /Root")
	}
	if !bytes.Contains(pdf, []byte("/Type /Catalog")) || !bytes.Contains(pdf, []byte("/Count 1")) {
		t.Errorf("want a catalog and a page tree with one page")
	}
	if n := bytes.Count(pdf, []byte("/Type /Page ")); n != 1 {
		t.Errorf("found %d page objects, want 1", n)
	}

	// Each stream's /Length is its exact byte count.
	streams := regexp.MustCompile(`(?s)/Length (\d+) >>\nstream\n(.*?)\nendstream`).FindAllSubmatch(pdf, -1)
	if len(streams) == 0 {
		t.Fatalf("no content stream found")
	}
	for _, s := range streams {
		n, _ := strconv.Atoi(string(s[1]))
		if n != len(s[2]) {
			t.Errorf("stream /Length %d, but it has %d bytes", n, len(s[2]))
		}
	}
}

func TestRenderContent(t *testing.T) {
	pdf, err := Render(sample())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	text := string(pdf)
	for _, want := range []string{
		"(TAX INVOICE)",
		"(Invoice No: INV-2026-0042)",
		"(Date: 24 Oct 2026)",
		"(Mama Njeri Hardware Ltd)",
		"(KRA PIN: P051234567Q)",
		"(KRA PIN: A123456789B)",
		`(Kamau \(Builders\) & Sons)`,
		`(Nails 3" \(1 kg\))`,
		"(Ksh 7,500.00)",
		"(Ksh 3,598.00)",
		"(Ksh 11,638.00)",
		"(VAT 16%)",
		"(Ksh 1,862.08)",
		"(Ksh 13,500.08)",
		"/BaseFont /Helvetica",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the PDF doesn't contain %s", want)
		}
	}
}
