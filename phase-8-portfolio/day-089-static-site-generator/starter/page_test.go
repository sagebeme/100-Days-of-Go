package ssg

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseFrontMatter(t *testing.T) {
	src := "---\ntitle: \"Matatu art: a short history\"\n# a comment\n\ndate: 2026-10-24\nTags: art, Nairobi\n---\n# Hello\n"
	meta, body, err := ParseFrontMatter([]byte(src))
	if err != nil {
		t.Fatalf("ParseFrontMatter: %v", err)
	}
	want := map[string]string{"title": "Matatu art: a short history", "date": "2026-10-24", "tags": "art, Nairobi"}
	if !reflect.DeepEqual(meta, want) {
		t.Errorf("meta = %v, want %v", meta, want)
	}
	if string(body) != "# Hello\n" {
		t.Errorf("body = %q, want %q", body, "# Hello\n")
	}
}

func TestParseFrontMatterWindowsLineEndings(t *testing.T) {
	meta, body, err := ParseFrontMatter([]byte("---\r\ntitle: Habari\r\n---\r\nNdiyo\r\n"))
	if err != nil || meta["title"] != "Habari" || string(body) != "Ndiyo\n" {
		t.Errorf("got %v, %q, %v; want title Habari and body \"Ndiyo\\n\"", meta, body, err)
	}
}

func TestParseFrontMatterErrors(t *testing.T) {
	tests := []struct {
		name, src string
		want      error // nil means any error
	}{
		{"none", "# Just markdown\n", ErrNoFrontMatter},
		{"not at the top", "\n---\ntitle: x\n---\n", ErrNoFrontMatter},
		{"unclosed", "---\ntitle: x\n# body\n", ErrUnclosed},
		{"no colon", "---\ntitle x\n---\n", nil},
		{"unknown key", "---\ntitel: x\n---\n", nil},
	}
	for _, tt := range tests {
		_, _, err := ParseFrontMatter([]byte(tt.src))
		if err == nil || (tt.want != nil && !errors.Is(err, tt.want)) {
			t.Errorf("%s: error = %v, want %v", tt.name, err, tt.want)
		}
	}
}

func TestSlugs(t *testing.T) {
	slugify := []struct{ in, want string }{
		{"hello.md", "hello"},
		{"Matatu Art 2026.md", "matatu-art-2026"},
		{"--why_go--.md", "why-go"},
		{"ugali & sukuma.md", "ugali-sukuma"},
		{"v1.2-notes.md", "v1-2-notes"},
	}
	for _, tt := range slugify {
		if got := Slugify(tt.in); got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
	valid := []struct {
		s    string
		want bool
	}{
		{"hello", true},
		{"matatu-art-2026", true},
		{"", false},
		{"Hello", false},
		{"a--b", false},
		{"-a", false},
		{"../etc", false},
		{"a/b", false},
	}
	for _, tt := range valid {
		if got := ValidSlug(tt.s); got != tt.want {
			t.Errorf("ValidSlug(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	md := []byte("# Karibu\n\nSome *bold* text.\n\n<script>alert(1)</script>\n\n[click](javascript:alert(1))\n\n| Route | Fare |\n|---|---|\n| 125 | Ksh100 |\n")

	safe, err := RenderMarkdown(md, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<h1>Karibu</h1>", "<em>bold</em>", "<table>", "<td>Ksh100</td>"} {
		if !strings.Contains(string(safe), want) {
			t.Errorf("safe output has no %s:\n%s", want, safe)
		}
	}
	for _, bad := range []string{"<script>", "javascript:"} {
		if strings.Contains(string(safe), bad) {
			t.Errorf("safe output contains %s:\n%s", bad, safe)
		}
	}

	unsafe, _ := RenderMarkdown(md, true)
	if !strings.Contains(string(unsafe), "<script>alert(1)</script>") {
		t.Errorf("with allowHTML, raw HTML should pass through:\n%s", unsafe)
	}
}

func TestParsePage(t *testing.T) {
	src := "---\ntitle: Ugali na Sukuma\ndescription: Dinner for Ksh 150\ndate: 2026-10-20\ndraft: false\ntags: Food, Nairobi,\n---\nCook the **ugali** first.\n"
	p, err := ParsePage("Ugali & Sukuma.md", []byte(src), false)
	if err != nil {
		t.Fatalf("ParsePage: %v", err)
	}
	want := Page{
		Title:       "Ugali na Sukuma",
		Description: "Dinner for Ksh 150",
		Date:        time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		Tags:        []string{"food", "nairobi"},
		Slug:        "ugali-sukuma",
		Body:        "<p>Cook the <strong>ugali</strong> first.</p>\n",
	}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("ParsePage =\n%+v\nwant\n%+v", p, want)
	}

	p, err = ParsePage("x.md", []byte("---\ntitle: X\nslug: my-post\ndraft: true\n---\n"), false)
	if err != nil || p.Slug != "my-post" || !p.Draft {
		t.Errorf("slug and draft from front matter: got %+v, %v", p, err)
	}
}

func TestParsePageErrors(t *testing.T) {
	tests := []struct {
		name, src string
		want      error
	}{
		{"no title", "---\ndate: 2026-10-24\n---\n", ErrNoTitle},
		{"path in slug", "---\ntitle: x\nslug: ../../etc\n---\n", ErrBadSlug},
		{"bad date", "---\ntitle: x\ndate: 24/10/2026\n---\n", nil},
		{"bad draft", "---\ntitle: x\ndraft: ndiyo\n---\n", nil},
	}
	for _, tt := range tests {
		_, err := ParsePage("post.md", []byte(tt.src), false)
		if err == nil || (tt.want != nil && !errors.Is(err, tt.want)) {
			t.Errorf("%s: error = %v, want %v", tt.name, err, tt.want)
			continue
		}
		if !strings.Contains(err.Error(), "post.md") {
			t.Errorf("%s: error %q doesn't name the file", tt.name, err)
		}
	}
}
