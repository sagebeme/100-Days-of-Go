package ssg

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Options control a build.
type Options struct {
	SiteTitle string
	Drafts    bool // include posts with draft: true
	AllowHTML bool // let raw HTML in Markdown through: only for authors you trust
}

// PageData is what the "page" template gets.
type PageData struct {
	Site Options
	Page Page
}

// IndexData is what the "index" template gets.
type IndexData struct {
	Site  Options
	Pages []Page
}

// ErrDuplicateSlug means two posts would be written to the same folder.
var ErrDuplicateSlug = errors.New("two posts have the same slug")

// Build reads a site from src and returns the files to write, keyed by slash-separated
// path. src holds:
//
//   - layout.html: html/template definitions named "page" and "index"
//   - posts/*.md: the posts (files in sub-folders and other extensions are ignored)
//   - static/...: copied as they are, under static/
//
// Each post goes to <slug>/index.html and the list of posts to index.html, newest
// first, then by slug. Drafts are left out unless opts.Drafts is set.
func Build(src fs.FS, opts Options) (map[string][]byte, error) {
	layout, err := fs.ReadFile(src, "layout.html")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("layout.html").Parse(string(layout))
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"page", "index"} {
		if tmpl.Lookup(name) == nil {
			return nil, fmt.Errorf("layout.html must define a template called %q", name)
		}
	}

	names, err := fs.Glob(src, "posts/*.md")
	if err != nil {
		return nil, err
	}
	var pages []Page
	seen := map[string]string{}
	for _, name := range names {
		b, err := fs.ReadFile(src, name)
		if err != nil {
			return nil, err
		}
		p, err := ParsePage(path.Base(name), b, opts.AllowHTML)
		if err != nil {
			return nil, err
		}
		if p.Draft && !opts.Drafts {
			continue
		}
		if other, ok := seen[p.Slug]; ok {
			return nil, fmt.Errorf("%w: %s and %s are both %q", ErrDuplicateSlug, other, name, p.Slug)
		}
		seen[p.Slug] = name
		pages = append(pages, p)
	}
	slices.SortFunc(pages, func(a, b Page) int {
		if c := b.Date.Compare(a.Date); c != 0 {
			return c
		}
		return strings.Compare(a.Slug, b.Slug)
	})

	out := map[string][]byte{}
	for _, p := range pages {
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "page", PageData{Site: opts, Page: p}); err != nil {
			return nil, fmt.Errorf("%s: %w", p.Slug, err)
		}
		out[p.Slug+"/index.html"] = buf.Bytes()
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "index", IndexData{Site: opts, Pages: pages}); err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}
	out["index.html"] = buf.Bytes()

	err = fs.WalkDir(src, "static", func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && p == "static" {
			return fs.SkipDir // no static folder is fine
		}
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(src, p)
		out[p] = b
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Write saves files under dir, making folders as needed. Every path must be local
// (no "..", not absolute), so a bad slug can never write outside dir.
func Write(dir string, files map[string][]byte) error {
	for name, b := range files {
		p := filepath.FromSlash(name)
		if !filepath.IsLocal(p) {
			return fmt.Errorf("refusing to write %q outside %s", name, dir)
		}
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
