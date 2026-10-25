package invoice

import (
	"bytes"
	"fmt"
	"strings"
)

// Page size: A4 in points (1/72 inch).
const (
	PageWidth  = 595
	PageHeight = 842
)

// Escape makes s safe inside a PDF literal string ( ... ): it puts a backslash before
// \, ( and ), and turns anything outside printable ASCII into ?. The standard fonts
// used here only cover that range reliably.
func Escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\\' || r == '(' || r == ')':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r < 32 || r > 126:
			b.WriteByte('?')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// helvetica holds Helvetica's widths, in 1/1000 of the font size, for the characters
// in amounts. It's enough to right-align a column of numbers.
var helvetica = map[rune]float64{
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556, '8': 556, '9': 556,
	',': 278, '.': 278, ' ': 278, '-': 333, 'K': 667, 's': 500, 'h': 556, '%': 889,
}

// TextWidth is the width of s in points in Helvetica at size, for the characters in
// the table above. Others count as 556, an average.
func TextWidth(s string, size float64) float64 {
	w := 0.0
	for _, r := range s {
		cw, ok := helvetica[r]
		if !ok {
			cw = 556
		}
		w += cw
	}
	return w * size / 1000
}

// content builds a page's content stream.
type content struct{ bytes.Buffer }

func (c *content) text(font string, size, x, y float64, s string) {
	fmt.Fprintf(c, "BT /%s %g Tf %.2f %.2f Td (%s) Tj ET\n", font, size, x, y, Escape(s))
}

// right draws s so that it ends at x.
func (c *content) right(font string, size, x, y float64, s string) {
	c.text(font, size, x-TextWidth(s, size), y, s)
}

func (c *content) line(x1, y1, x2, y2 float64) {
	fmt.Fprintf(c, "0.5 w %.2f %.2f m %.2f %.2f l S\n", x1, y1, x2, y2)
}

// shorten cuts s to at most n runes, ending in "..." when it had to cut, so a long
// description can't run into the Qty column.
func shorten(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-3]) + "..."
}

// Render validates inv and returns a one-page PDF 1.4 file: a catalog, a page tree,
// one A4 page, two fonts (Helvetica and Helvetica-Bold) and a content stream, then
// the xref table, the trailer, startxref and %%EOF.
func Render(inv Invoice) ([]byte, error) {
	if err := inv.Validate(); err != nil {
		return nil, err
	}
	return assemble(page(inv)), nil
}

// page lays out the invoice.
func page(inv Invoice) []byte {
	const left, right = 50.0, 545.0
	var c content

	c.text("F2", 20, left, 780, "TAX INVOICE")
	c.right("F1", 10, right, 786, "Invoice No: "+inv.Number)
	c.right("F1", 10, right, 772, "Date: "+inv.Date.Format("2 Jan 2006"))

	y := 740.0
	c.text("F2", 12, left, y, inv.Seller.Name)
	for _, s := range []string{inv.Seller.Address, inv.Seller.Phone, "KRA PIN: " + inv.Seller.KRAPIN} {
		if strings.TrimSpace(s) != "" {
			y -= 14
			c.text("F1", 10, left, y, s)
		}
	}

	y -= 30
	c.text("F2", 10, left, y, "Bill to:")
	buyer := []string{inv.Buyer.Name, inv.Buyer.Address, inv.Buyer.Phone}
	if inv.Buyer.KRAPIN != "" {
		buyer = append(buyer, "KRA PIN: "+inv.Buyer.KRAPIN)
	}
	for _, s := range buyer {
		if strings.TrimSpace(s) != "" {
			y -= 14
			c.text("F1", 10, left, y, s)
		}
	}

	const qtyX, unitX = 360.0, 450.0
	y -= 36
	c.text("F2", 10, left, y, "Description")
	c.right("F2", 10, qtyX, y, "Qty")
	c.right("F2", 10, unitX, y, "Unit price")
	c.right("F2", 10, right, y, "Amount")
	c.line(left, y-6, right, y-6)
	for _, l := range inv.Lines {
		y -= 20
		c.text("F1", 10, left, y, shorten(l.Description, 48))
		c.right("F1", 10, qtyX, y, fmt.Sprint(l.Qty))
		c.right("F1", 10, unitX, y, FormatKsh(l.UnitCents))
		c.right("F1", 10, right, y, FormatKsh(l.Amount()))
	}
	c.line(left, y-8, right, y-8)

	t := ComputeTotals(inv.Lines)
	rows := []struct {
		label string
		cents int64
		font  string
	}{
		{"Subtotal", t.Subtotal, "F1"},
		{fmt.Sprintf("VAT %d%%", VATPercent), t.VAT, "F1"},
		{"Total", t.Total, "F2"},
	}
	for _, r := range rows {
		y -= 20
		c.right(r.font, 10, unitX, y, r.label)
		c.right(r.font, 10, right, y, FormatKsh(r.cents))
	}

	c.text("F1", 8, left, 60, "Amounts in Kenya shillings. Prices exclude VAT; VAT is charged at 16% on the subtotal.")
	return c.Bytes()
}

// assemble wraps a content stream in the objects a PDF needs, then writes the
// cross-reference table and the trailer.
func assemble(stream []byte) []byte {
	var buf bytes.Buffer
	var offsets []int // offsets[i] is where object i+1 starts

	// The second line has bytes above 127, telling tools the file is binary.
	buf.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	obj := func(body string) {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", len(offsets), body)
	}

	obj("<< /Type /Catalog /Pages 2 0 R >>")
	obj("<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	obj(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] "+
		"/Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>", PageWidth, PageHeight))
	obj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	obj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	obj(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))

	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(offsets)+1)
	buf.WriteString("0000000000 65535 f \n") // each entry is exactly 20 bytes
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xref)
	return buf.Bytes()
}
