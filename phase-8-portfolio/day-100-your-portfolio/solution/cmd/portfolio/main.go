// Command portfolio reads projects.json and writes index.html, BADGES.md and CV.md.
//
//	go run ./phase-8-portfolio/day-100-your-portfolio/solution/cmd/portfolio -in projects.json -out site
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-100-your-portfolio/solution"
)

func main() {
	in := flag.String("in", "projects.json", "the projects file to read")
	out := flag.String("out", "site", "the folder to write index.html, BADGES.md and CV.md into")
	flag.Parse()
	if err := run(*in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "portfolio:", err)
		os.Exit(1)
	}
}

func run(in, out string) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()
	p, err := portfolio.Load(f)
	if err != nil {
		return err
	}
	var page bytes.Buffer
	if err := portfolio.RenderPage(&page, p); err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	files := []struct {
		name string
		data []byte
	}{
		{"index.html", page.Bytes()},
		{"BADGES.md", []byte(portfolio.BadgeTable(p))},
		{"CV.md", []byte(portfolio.CVLines(p))},
	}
	for _, f := range files {
		path := filepath.Join(out, f.name)
		if err := os.WriteFile(path, f.data, 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}
	return nil
}
