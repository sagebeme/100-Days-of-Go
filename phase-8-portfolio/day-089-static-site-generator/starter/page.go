// Package ssg is a static site generator: Markdown posts with front matter, rendered
// by goldmark into html/template layouts, plus an index page.
package ssg

import (
	"errors"
	"html/template"
	"time"
)

// Page is one post.
type Page struct {
	Title       string
	Description string
	Date        time.Time
	Draft       bool
	Tags        []string
	Slug        string        // the URL is /<slug>/
	Body        template.HTML // the rendered Markdown
}

// Errors from ParseFrontMatter and ParsePage.
var (
	ErrNoFrontMatter = errors.New("a post must start with --- front matter")
	ErrUnclosed      = errors.New("front matter has no closing ---")
	ErrNoTitle       = errors.New("front matter needs a title")
	ErrBadSlug       = errors.New("slug must be lower-case letters, digits and hyphens")
)

// Keys lists the front matter keys a post may use. Anything else is an error, which
// catches typos like "titel".
var Keys = []string{"title", "description", "date", "draft", "tags", "slug"}

// ParseFrontMatter splits a post into its front matter and its Markdown body. The
// post must start with a line "---" and the front matter ends at the next "---" line.
// Each line in between is "key: value"; blank lines and lines starting with # are
// skipped. Keys are lower-cased and must be in Keys. Values are trimmed, and one pair
// of surrounding double quotes is removed. Windows line endings are accepted.
func ParseFrontMatter(src []byte) (map[string]string, []byte, error) {
	// Your code here
	return nil, nil, nil
}

// ValidSlug reports whether s is safe to use as a folder name and a URL: lower-case
// letters and digits in groups joined by single hyphens.
func ValidSlug(s string) bool {
	// Your code here
	return false
}

// Slugify turns a file name like "Matatu Art 2026.md" into "matatu-art-2026": it drops
// the extension, lower-cases, and turns every run of other characters into one hyphen.
func Slugify(name string) string {
	// Your code here
	return ""
}

// RenderMarkdown turns Markdown into HTML with GitHub-flavoured extensions (tables,
// strikethrough, autolinks, task lists). Unless allowHTML is true, raw HTML in the
// Markdown is left out and javascript: links lose their href.
func RenderMarkdown(md []byte, allowHTML bool) (template.HTML, error) {
	// Your code here
	return "", nil
}

// ParsePage reads one post. name is its file name, used for the slug when the front
// matter has none. date is YYYY-MM-DD, draft is true or false, and tags are separated
// by commas and lower-cased. Errors mention name.
func ParsePage(name string, src []byte, allowHTML bool) (Page, error) {
	// Your code here
	return Page{}, nil
}
