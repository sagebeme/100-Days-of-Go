// Command invoice writes a sample invoice to a PDF file.
//
//	go run ./phase-8-portfolio/day-088-invoice-pdf/solution/cmd/invoice -o invoice.pdf
package main

import (
	"flag"
	"log"
	"os"
	"time"

	invoice "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-088-invoice-pdf/solution"
)

func main() {
	out := flag.String("o", "invoice.pdf", "file to write")
	flag.Parse()

	inv := invoice.Invoice{
		Number: "INV-2026-0042",
		Date:   time.Now(),
		Seller: invoice.Party{Name: "Mama Njeri Hardware Ltd", KRAPIN: "P051234567Q", Address: "Moi Avenue, Nairobi", Phone: "0712 345 678"},
		Buyer:  invoice.Party{Name: "Kamau Builders", Address: "Ngong Road, Nairobi"},
		Lines: []invoice.Line{
			{Description: "Cement, 50 kg bag", Qty: 10, UnitCents: 75000},
			{Description: "Iron sheets, gauge 30", Qty: 4, UnitCents: 89950},
			{Description: "Nails 3 inch (1 kg)", Qty: 3, UnitCents: 18000},
		},
	}
	pdf, err := invoice.Render(inv)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, pdf, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d bytes)", *out, len(pdf))
}
