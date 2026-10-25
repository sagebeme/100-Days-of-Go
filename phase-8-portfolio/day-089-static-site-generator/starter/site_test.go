package ssg

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

const layout = `{{define "page"}}<title>{{.Page.Title}} | {{.Site.SiteTitle}}</title>
<h1>{{.Page.Title}}</h1>
<time>{{.Page.Date.Format "2 Jan 2006"}}</time>
{{.Page.Body}}{{end}}
{{define "index"}}<title>{{.Site.SiteTitle}}</title>
<ul>{{range .Pages}}
<li><a href="/{{.Slug}}/">{{.Title}}</a></li>{{end}}
</ul>{{end}}`

func site() fstest.MapFS {
	return fstest.MapFS{
		"layout.html":         {Data: []byte(layout)},
		"posts/older.md":      {Data: []byte("---\ntitle: Older post\ndate: 2026-01-05\n---\nFirst.\n")},
		"posts/newer.md":      {Data: []byte("---\ntitle: Kamau <script>alert(1)</script>\ndate: 2026-10-24\n---\nSecond <b>bold</b>.\n")},
		"posts/same-day-a.md": {Data: []byte("---\ntitle: Same day A\ndate: 2026-03-01\n---\nA\n")},
		"posts/same-day-b.md": {Data: []byte("---\ntitle: Same day B\ndate: 2026-03-01\n---\nB\n")},
		"posts/wip.md":        {Data: []byte("---\ntitle: Work in progress\ndate: 2026-10-25\ndraft: true\n---\nNot yet.\n")},
		"posts/notes.txt":     {Data: []byte("not a post")},
		"posts/sub/deep.md":   {Data: []byte("---\ntitle: In a sub-folder\n---\n")},
		"static/style.css":    {Data: []byte("body { font-family: sans-serif; }")},
		"static/img/logo.svg": {Data: []byte("<svg/>")},
	}
}

func keys(m map[string][]byte) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func TestBuildFiles(t *testing.T) {
	files, err := Build(site(), Options{SiteTitle: "Kamau's Blog"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := []string{
		"index.html",
		"newer/index.html",
		"older/index.html",
		"same-day-a/index.html",
		"same-day-b/index.html",
		"static/img/logo.svg",
		"static/style.css",
	}
	if got := keys(files); !slices.Equal(got, want) {
		t.Errorf("files =\n%v\nwant\n%v", got, want)
	}
	if got := string(files["static/style.css"]); got != "body { font-family: sans-serif; }" {
		t.Errorf("static file changed: %q", got)
	}
}

func TestBuildIndexOrder(t *testing.T) {
	files, err := Build(site(), Options{SiteTitle: "Kamau's Blog"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	index := string(files["index.html"])
	order := []string{`href="/newer/"`, `href="/same-day-a/"`, `href="/same-day-b/"`, `href="/older/"`}
	last := -1
	for _, link := range order {
		i := strings.Index(index, link)
		if i < 0 {
			t.Fatalf("index has no %s:\n%s", link, index)
		}
		if i < last {
			t.Errorf("index lists %s too early: want newest first, then by slug", link)
		}
		last = i
	}
	if !strings.Contains(index, "<title>Kamau&#39;s Blog</title>") {
		t.Errorf("index title not escaped by html/template:\n%s", index)
	}
}

func TestBuildEscapes(t *testing.T) {
	files, err := Build(site(), Options{SiteTitle: "Blog"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	page := string(files["newer/index.html"])
	if strings.Contains(page, "<script>") || strings.Contains(page, "<b>") {
		t.Errorf("raw HTML got through:\n%s", page)
	}
	if !strings.Contains(page, "<h1>Kamau &lt;script&gt;alert(1)&lt;/script&gt;</h1>") {
		t.Errorf("the title should be escaped by html/template:\n%s", page)
	}
	if !strings.Contains(page, "<time>24 Oct 2026</time>") {
		t.Errorf("page has no date:\n%s", page)
	}

	files, _ = Build(site(), Options{SiteTitle: "Blog", AllowHTML: true})
	if page := string(files["newer/index.html"]); !strings.Contains(page, "<b>bold</b>") {
		t.Errorf("with AllowHTML, raw HTML in the body should pass:\n%s", page)
	}
}

func TestBuildDrafts(t *testing.T) {
	files, err := Build(site(), Options{Drafts: true})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, ok := files["wip/index.html"]; !ok {
		t.Errorf("with Drafts, wip/index.html is missing")
	}
	if !strings.Contains(string(files["index.html"]), "Work in progress") {
		t.Errorf("with Drafts, the index should list the draft")
	}
}

func TestBuildErrors(t *testing.T) {
	dup := site()
	dup["posts/copy.md"] = &fstest.MapFile{Data: []byte("---\ntitle: Copy\nslug: older\n---\n")}
	if _, err := Build(dup, Options{}); !errors.Is(err, ErrDuplicateSlug) {
		t.Errorf("duplicate slug: error = %v, want ErrDuplicateSlug", err)
	}

	bad := site()
	bad["posts/broken.md"] = &fstest.MapFile{Data: []byte("no front matter")}
	if _, err := Build(bad, Options{}); !errors.Is(err, ErrNoFrontMatter) || !strings.Contains(err.Error(), "broken.md") {
		t.Errorf("broken post: error = %v, want ErrNoFrontMatter naming broken.md", err)
	}

	noLayout := site()
	delete(noLayout, "layout.html")
	if _, err := Build(noLayout, Options{}); err == nil {
		t.Errorf("no layout.html: want an error")
	}

	noIndex := site()
	noIndex["layout.html"] = &fstest.MapFile{Data: []byte(`{{define "page"}}x{{end}}`)}
	if _, err := Build(noIndex, Options{}); err == nil {
		t.Errorf("a layout without an index template: want an error")
	}

	noStatic := site()
	delete(noStatic, "static/style.css")
	delete(noStatic, "static/img/logo.svg")
	if _, err := Build(noStatic, Options{}); err != nil {
		t.Errorf("a site with no static folder: %v, want no error", err)
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{"index.html": []byte("home"), "a/b/index.html": []byte("deep")}
	if err := Write(dir, files); err != nil {
		t.Fatalf("Write: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "a", "b", "index.html"))
	if err != nil || string(b) != "deep" {
		t.Errorf("a/b/index.html = %q, %v", b, err)
	}

	for _, bad := range []string{"../escape.html", "/etc/passwd"} {
		if err := Write(dir, map[string][]byte{bad: []byte("x")}); err == nil {
			t.Errorf("Write(%q) succeeded, want an error", bad)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "escape.html")); err == nil {
		t.Errorf("Write created a file outside the folder")
	}
}
