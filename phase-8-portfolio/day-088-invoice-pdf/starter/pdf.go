package invoice

// Page size: A4 in points (1/72 inch).
const (
	PageWidth  = 595
	PageHeight = 842
)

// Escape makes s safe inside a PDF literal string ( ... ): it puts a backslash before
// \, ( and ), and turns anything outside printable ASCII into ?. The standard fonts
// used here only cover that range reliably.
func Escape(s string) string {
	// Your code here
	return ""
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
	// Your code here
	return 0
}

// Render validates inv and returns a one-page PDF 1.4 file: a catalog, a page tree,
// one A4 page, two fonts (Helvetica and Helvetica-Bold) and a content stream, then
// the xref table, the trailer, startxref and %%EOF.
func Render(inv Invoice) ([]byte, error) {
	// Your code here
	return nil, nil
}
