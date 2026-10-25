// Command ssg builds a site folder into HTML, and can serve the result to preview it.
//
//	go run ./phase-8-portfolio/day-089-static-site-generator/starter/cmd/ssg \
//	    -src phase-8-portfolio/day-089-static-site-generator/example -out public -serve :8080
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	ssg "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-089-static-site-generator/starter"
)

func main() {
	src := flag.String("src", ".", "site folder with layout.html, posts/ and static/")
	out := flag.String("out", "public", "folder to write the site to")
	title := flag.String("title", "My Blog", "site title")
	drafts := flag.Bool("drafts", false, "include draft posts")
	allowHTML := flag.Bool("html", false, "let raw HTML in Markdown through")
	serve := flag.String("serve", "", "after building, serve -out on this address, e.g. :8080")
	flag.Parse()

	files, err := ssg.Build(os.DirFS(*src), ssg.Options{SiteTitle: *title, Drafts: *drafts, AllowHTML: *allowHTML})
	if err != nil {
		log.Fatal(err)
	}
	if err := ssg.Write(*out, files); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %d files to %s", len(files), *out)

	if *serve != "" {
		log.Printf("serving %s on %s", *out, *serve)
		srv := &http.Server{Addr: *serve, Handler: http.FileServer(http.Dir(*out)), ReadHeaderTimeout: 5 * time.Second}
		log.Fatal(srv.ListenAndServe())
	}
}
