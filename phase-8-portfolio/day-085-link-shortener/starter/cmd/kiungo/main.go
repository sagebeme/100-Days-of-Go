// Command kiungo runs the link shortener with an in-memory store.
//
//	go run ./phase-8-portfolio/day-085-link-shortener/starter/cmd/kiungo -addr :8080
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	linkshortener "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-085-link-shortener/starter"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	base := flag.String("base", "http://localhost:8080", "public base URL for short links")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           linkshortener.New(linkshortener.Config{Store: linkshortener.NewMemoryStore(), BaseURL: *base}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("kiungo listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
