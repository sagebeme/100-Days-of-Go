// Package ssg is a static site generator: Markdown posts with front matter, rendered
// by goldmark into html/template layouts, plus an index page.
package ssg

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
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
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return nil, nil, ErrNoFrontMatter
	}
	meta := map[string]string{}
	for n := 2; ; n++ {
		line, after, found := strings.Cut(rest, "\n")
		if !found && line != "---" {
			return nil, nil, ErrUnclosed
		}
		rest = after
		if line == "---" {
			return meta, []byte(rest), nil
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return nil, nil, fmt.Errorf("line %d: want key: value, got %q", n, line)
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		if !isKey(key) {
			return nil, nil, fmt.Errorf("line %d: unknown key %q (want one of %s)", n, key, strings.Join(Keys, ", "))
		}
		meta[key] = value
	}
}

func isKey(k string) bool {
	for _, want := range Keys {
		if k == want {
			return true
		}
	}
	return false
}

var slugRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidSlug reports whether s is safe to use as a folder name and a URL: lower-case
// letters and digits in groups joined by single hyphens.
func ValidSlug(s string) bool { return slugRE.MatchString(s) }

// Slugify turns a file name like "Matatu Art 2026.md" into "matatu-art-2026": it drops
// the extension, lower-cases, and turns every run of other characters into one hyphen.
func Slugify(name string) string {
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		name = name[:i]
	}
	var b strings.Builder
	hyphen := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if hyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			hyphen = false
			b.WriteRune(r)
		} else {
			hyphen = true
		}
	}
	return b.String()
}

// RenderMarkdown turns Markdown into HTML with GitHub-flavoured extensions (tables,
// strikethrough, autolinks, task lists). Unless allowHTML is true, raw HTML in the
// Markdown is left out and javascript: links lose their href.
func RenderMarkdown(md []byte, allowHTML bool) (template.HTML, error) {
	opts := []goldmark.Option{goldmark.WithExtensions(extension.GFM)}
	if allowHTML {
		opts = append(opts, goldmark.WithRendererOptions(html.WithUnsafe()))
	}
	var buf bytes.Buffer
	if err := goldmark.New(opts...).Convert(md, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// ParsePage reads one post. name is its file name, used for the slug when the front
// matter has none. date is YYYY-MM-DD, draft is true or false, and tags are separated
// by commas and lower-cased. Errors mention name.
func ParsePage(name string, src []byte, allowHTML bool) (Page, error) {
	meta, body, err := ParseFrontMatter(src)
	if err != nil {
		return Page{}, fmt.Errorf("%s: %w", name, err)
	}
	p := Page{Title: meta["title"], Description: meta["description"], Slug: meta["slug"]}
	if p.Title == "" {
		return Page{}, fmt.Errorf("%s: %w", name, ErrNoTitle)
	}
	if p.Slug == "" {
		p.Slug = Slugify(name)
	}
	if !ValidSlug(p.Slug) {
		return Page{}, fmt.Errorf("%s: %w: %q", name, ErrBadSlug, p.Slug)
	}
	if d := meta["date"]; d != "" {
		if p.Date, err = time.Parse(time.DateOnly, d); err != nil {
			return Page{}, fmt.Errorf("%s: date must be YYYY-MM-DD: %q", name, d)
		}
	}
	if d := meta["draft"]; d != "" {
		if p.Draft, err = strconv.ParseBool(d); err != nil {
			return Page{}, fmt.Errorf("%s: draft must be true or false: %q", name, d)
		}
	}
	for _, t := range strings.Split(meta["tags"], ",") {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
			p.Tags = append(p.Tags, t)
		}
	}
	if p.Body, err = RenderMarkdown(body, allowHTML); err != nil {
		return Page{}, fmt.Errorf("%s: %w", name, err)
	}
	return p, nil
}
