package ssg

import (
	"errors"
	"io/fs"
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
	// Your code here
	return nil, nil
}

// Write saves files under dir, making folders as needed. Every path must be local
// (no "..", not absolute), so a bad slug can never write outside dir.
func Write(dir string, files map[string][]byte) error {
	// Your code here
	return nil
}
