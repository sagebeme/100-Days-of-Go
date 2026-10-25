# Day 88: Invoice PDF

## The brief

Njeri runs Mama Njeri Hardware on Moi Avenue. Contractors keep asking for "a proper invoice, a PDF, with your KRA PIN on it", and she's been typing them in a word processor and getting the VAT sums wrong. She wants a small Go program that turns an invoice's data into a one-page PDF: her details and KRA PIN, the buyer, the lines, the subtotal, VAT at 16% and the total in Ksh. Her nephew says "just use a PDF library". You'll write the PDF by hand instead, so you know exactly what's in the file.

```
ComputeTotals(lines)    →  {Subtotal: 1163800, VAT: 186208, Total: 1350008}   (cents)
FormatKsh(1350008)      →  "Ksh 13,500.08"
ValidKRAPIN("P051234567Q") → true
Escape("Kamau (Builders)") →  `Kamau \(Builders\)`
Render(inv)             →  %PDF-1.4 ... xref ... trailer ... startxref 2335 %%EOF
Render(inv with no PIN) →  error: seller KRA PIN must look like P051234567Q
```

```
TAX INVOICE                                           Invoice No: INV-2026-0042
Mama Njeri Hardware Ltd                                       Date: 24 Oct 2026
KRA PIN: P051234567Q
Description                      Qty     Unit price         Amount
Cement, 50 kg bag                 10     Ksh 750.00   Ksh 7,500.00
...
                                           Subtotal  Ksh 11,638.00
                                            VAT 16%   Ksh 1,862.08
                                              Total  Ksh 13,500.08
```

Acceptance criteria:

- Money is whole cents in `int64`, never `float64`. VAT is 16% of the subtotal, rounded to the nearest cent.
- The seller's KRA PIN is required and checked for format. The buyer's is optional.
- The output opens in any PDF reader: correct header, objects, a cross-reference table whose offsets point at each object exactly, a trailer and `%%EOF`.
- Text from the invoice can't break the PDF: `(`, `)` and `\` are escaped.

A note on the law: Kenyan businesses are required to issue tax invoices through KRA's eTIMS system. Treat today's PDF as practice, or as the quote a business sends before the eTIMS invoice.

## What you'll use

- Money as whole numbers: [Day 2](../../phase-1-terminal/day-002-bundle-calculator/). Formatting columns of amounts: [Day 11](../../phase-1-terminal/day-011-receipt-printer/).
- Checking input byte by byte, for the KRA PIN: [Day 9](../../phase-1-terminal/day-009-phone-cleaner/).
- Sentinel errors: [Day 17](../../phase-2-modelling/day-017-failed-transactions/). Table tests: [Day 19](../../phase-2-modelling/day-019-table-tests/).
- `bytes.Buffer`, whose `Len()` is the byte offset of whatever you write next. That's all an xref table needs.

A PDF is text with byte offsets. The smallest useful one:

```
%PDF-1.4
1 0 obj  << /Type /Catalog /Pages 2 0 R >>  endobj
2 0 obj  << /Type /Pages /Kids [3 0 R] /Count 1 >>  endobj
3 0 obj  << /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R ... >>  endobj
4 0 obj  << /Length 42 >> stream
BT /F1 20 Tf 50 780 Td (TAX INVOICE) Tj ET
endstream endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n     ← byte offset of "1 0 obj"
...
trailer << /Size 5 /Root 1 0 R >>
startxref
412                    ← byte offset of "xref"
%%EOF
```

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `Line.Amount`, `ComputeTotals` and `FormatKsh` with thousands separators and negatives.
2. `ValidKRAPIN`: `A` or `P`, nine digits, one capital letter.
3. `Validate` returns the first problem, in the order the errors are listed, including overflow checks.
4. `Escape` and `TextWidth`. Use `TextWidth` to right-align the Qty, price and amount columns.
5. `Render` writes a content stream with `BT ... Tj ET` for text and `m ... l S` for the table rules, then six objects: catalog, pages, page, two Type1 fonts and the stream. Record each object's offset as you write it.
6. The xref table: `0 N`, the free entry `0000000000 65535 f `, then one entry per object, each exactly 20 bytes. Then `trailer`, `/Size N /Root 1 0 R`, `startxref` with the table's offset, and `%%EOF`.
7. A stream's `/Length` is its exact byte count, not counting the newline before `endstream`.
8. `cmd/invoice` writes a sample to a file. Open it in a PDF reader, and check it with `pdfinfo` or `qpdf --check` if you have them.

```bash
go test ./phase-8-portfolio/day-088-invoice-pdf/starter
go run ./phase-8-portfolio/day-088-invoice-pdf/starter/cmd/invoice -o invoice.pdf
```

Stretch goals:

- Read the invoice from a JSON file ([Day 23](../../phase-2-modelling/day-023-save-load/)) and add `-o` and `-in` flags.
- Per-line VAT classes: standard 16%, zero-rated and exempt, with a VAT summary by class.
- Compress the content stream with `compress/zlib` and `/Filter /FlateDecode`. The tests on structure should still pass.
- Several pages when the lines don't fit, with "Page 1 of 2" and the totals on the last page.
- Serve it over HTTP: `POST /invoices` returns `application/pdf` with `Content-Disposition: attachment`.

## The kata: Xref Table

Given a header size and the body length of each object, print the xref table and `startxref`. It's the bookkeeping inside `Render`, where the digit count of the object number is the trap. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 88
```

## When you're stuck

- **The reader says the file is damaged, or `pdfinfo` prints `Syntax Error: Couldn't find trailer dictionary` or warns about the xref**: an offset is off. Record `buf.Len()` immediately before writing `N 0 obj`, not after.
- **Every offset is off by the same few bytes**: the header. `%PDF-1.4\n` is 9 bytes, and the binary comment line after it counts too.
- **The xref test says an entry isn't 20 bytes**: the format is `%010d 00000 n` plus a space and `\n`. Without the space, it's 19.
- **The text is missing but the page opens**: the font isn't in the page's `/Resources`, or the stream `/Length` is wrong. Count bytes with `len(stream)`.
- **VAT is `186207` instead of `186208`**: you used `float64`, or truncated. Add 50 before dividing by 100.
- **`Kamau (Builders)` breaks every line after it**: an unescaped `)` closes the string early. Escape `\` first or together with the brackets, in one pass.
- **`Ksh 1160.00` with no comma**: insert a comma before every group of three digits counted from the right, not the left.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Give it its own repo with a README that explains the PDF structure in a diagram (header, objects, xref, trailer) and shows a screenshot of a rendered invoice.
- Include a sample `invoice.pdf` in the repo, generated by CI, so readers can open it without installing Go.
- Add a small web form that posts JSON and downloads the PDF, and deploy it with a demo link.
- Write "What I'd do next": several pages, VAT classes, compressed streams, eTIMS integration.
- Add GitHub Actions running `go test`, plus `qpdf --check` on the generated file, and show the CI badge.
