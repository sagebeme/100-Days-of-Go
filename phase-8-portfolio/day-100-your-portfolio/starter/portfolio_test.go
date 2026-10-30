package portfolio

import (
	"slices"
	"strings"
	"testing"
)

const good = `{
  "owner": "Amina Wanjiru",
  "headline": "Junior Go developer in Nairobi. I build backends that don't lose money.",
  "projects": [
    {"name": "Key-Value Store", "day": 93, "summary": "A tiny database with a write-ahead log that survives a power cut.",
     "repo": "https://github.com/aminaw/kvstore", "tags": ["go", "storage"],
     "cv_line": "Built a crash-safe key-value store in Go with a CRC-checked write-ahead log and compaction."},
    {"name": "Webhook Relay", "day": 95, "summary": "Verifies, dedupes and forwards payment webhooks with retries.",
     "repo": "https://github.com/aminaw/relay", "demo": "https://relay.example.dev", "tags": ["go", "security"], "featured": true,
     "cv_line": "Built an HMAC-verified webhook relay with replay protection and exponential backoff."},
    {"name": "Job Board", "day": 94, "summary": "Search, pagination and server-rendered pages in plain Go.",
     "repo": "https://github.com/aminaw/kazi", "demo": "https://kazi.example.dev",
     "cv_line": "Shipped a job board with ranked search and CSRF protection, deployed on a free tier."}
  ]
}`

func load(t *testing.T, s string) Portfolio {
	t.Helper()
	p, err := Load(strings.NewReader(s))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return p
}

func TestLoad(t *testing.T) {
	p := load(t, good)
	if p.Owner != "Amina Wanjiru" || len(p.Projects) != 3 {
		t.Fatalf("Load = owner %q, %d projects; want Amina Wanjiru and 3", p.Owner, len(p.Projects))
	}
	relay := p.Projects[1]
	if relay.Day != 95 || !relay.Featured || relay.Demo != "https://relay.example.dev" || !slices.Equal(relay.Tags, []string{"go", "security"}) {
		t.Errorf("second project = %+v", relay)
	}
	if _, err := Load(strings.NewReader(`{"owner": "A", "projects": [], "colour": "blue"}`)); err == nil {
		t.Error("Load accepted an unknown field, want an error (a typo in a field name must not pass silently)")
	}
	if _, err := Load(strings.NewReader(`{"owner": `)); err == nil {
		t.Error("Load accepted broken JSON")
	}
}

func TestValidate(t *testing.T) {
	ok := func() Project {
		return Project{Name: "Relay", Day: 95, Summary: "Forwards webhooks.", Repo: "https://github.com/a/relay", CVLine: "Built a relay."}
	}
	tests := []struct {
		name   string
		change func(p *Portfolio)
		wantIn []string // each must appear in the error; nil means no error
	}{
		{"valid", func(p *Portfolio) {}, nil},
		{"no owner", func(p *Portfolio) { p.Owner = " " }, []string{"owner"}},
		{"no projects", func(p *Portfolio) { p.Projects = nil }, []string{"projects"}},
		{"day 0", func(p *Portfolio) { p.Projects[0].Day = 0 }, []string{"day"}},
		{"day 101", func(p *Portfolio) { p.Projects[0].Day = 101 }, []string{"day"}},
		{"long summary", func(p *Portfolio) { p.Projects[0].Summary = strings.Repeat("a", MaxSummary+1) }, []string{"summary"}},
		{"160 runes of Kiswahili is fine", func(p *Portfolio) { p.Projects[0].Summary = strings.Repeat("é", MaxSummary) }, nil},
		{"gitlab repo", func(p *Portfolio) { p.Projects[0].Repo = "https://gitlab.com/a/relay" }, []string{"repo"}},
		{"http repo", func(p *Portfolio) { p.Projects[0].Repo = "http://github.com/a/relay" }, []string{"repo"}},
		{"repo with a path", func(p *Portfolio) { p.Projects[0].Repo = "https://github.com/a/relay/tree/main" }, []string{"repo"}},
		{"just a user", func(p *Portfolio) { p.Projects[0].Repo = "https://github.com/a" }, []string{"repo"}},
		{"javascript demo", func(p *Portfolio) { p.Projects[0].Demo = "javascript:alert(1)" }, []string{"demo"}},
		{"http demo", func(p *Portfolio) { p.Projects[0].Demo = "http://relay.example.dev" }, []string{"demo"}},
		{"no cv line", func(p *Portfolio) { p.Projects[0].CVLine = "" }, []string{"cv_line"}},
		{"same name twice", func(p *Portfolio) { p.Projects = append(p.Projects, ok()); p.Projects[1].Name = "relay" }, []string{"name"}},
		{"several problems at once", func(p *Portfolio) { p.Owner = ""; p.Projects[0].Day = 0; p.Projects[0].Demo = "ftp://x" },
			[]string{"owner", "day", "demo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Portfolio{Owner: "Amina", Projects: []Project{ok()}}
			tt.change(&p)
			err := Validate(p)
			if tt.wantIn == nil {
				if err != nil {
					t.Fatalf("Validate = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate = nil, want an error mentioning %v", tt.wantIn)
			}
			for _, s := range tt.wantIn {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("Validate = %q, want it to mention %q", err, s)
				}
			}
		})
	}
}

func TestRepoPath(t *testing.T) {
	owner, repo, err := RepoPath("https://github.com/aminaw/relay")
	if err != nil || owner != "aminaw" || repo != "relay" {
		t.Errorf("RepoPath = %q, %q, %v; want aminaw, relay, nil", owner, repo, err)
	}
	if _, _, err := RepoPath("https://github.com/aminaw/relay.git"); err == nil {
		t.Error("RepoPath accepted a .git clone URL; want the web URL")
	}
}

func TestSorted(t *testing.T) {
	in := []Project{
		{Name: "B", Day: 90},
		{Name: "A", Day: 90},
		{Name: "C", Day: 99},
		{Name: "D", Day: 14, Featured: true},
		{Name: "E", Day: 50, Featured: true},
	}
	before := slices.Clone(in)
	var names []string
	for _, p := range Sorted(in) {
		names = append(names, p.Name)
	}
	if want := []string{"E", "D", "C", "A", "B"}; !slices.Equal(names, want) {
		t.Errorf("Sorted order = %v, want %v (featured, then latest day, then name)", names, want)
	}
	if !slices.EqualFunc(in, before, func(a, b Project) bool { return a.Name == b.Name }) {
		t.Error("Sorted changed its argument")
	}
}

func TestBadgeTable(t *testing.T) {
	got := BadgeTable(load(t, good))
	want := "| Project | Day | CI | Demo |\n" +
		"|---|---|---|---|\n" +
		"| [Webhook Relay](https://github.com/aminaw/relay) | 95 | [![CI](https://github.com/aminaw/relay/actions/workflows/ci.yml/badge.svg)](https://github.com/aminaw/relay/actions/workflows/ci.yml) | [demo](https://relay.example.dev) |\n" +
		"| [Job Board](https://github.com/aminaw/kazi) | 94 | [![CI](https://github.com/aminaw/kazi/actions/workflows/ci.yml/badge.svg)](https://github.com/aminaw/kazi/actions/workflows/ci.yml) | [demo](https://kazi.example.dev) |\n" +
		"| [Key-Value Store](https://github.com/aminaw/kvstore) | 93 | [![CI](https://github.com/aminaw/kvstore/actions/workflows/ci.yml/badge.svg)](https://github.com/aminaw/kvstore/actions/workflows/ci.yml) | - |\n"
	if got != want {
		t.Errorf("BadgeTable\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestBadgeTableEscapesPipes(t *testing.T) {
	p := Portfolio{Owner: "A", Projects: []Project{{Name: "Cache | Redis", Day: 78, Repo: "https://github.com/a/cache"}}}
	if got := BadgeTable(p); !strings.Contains(got, `[Cache \| Redis]`) {
		t.Errorf("a | in a name must be escaped as \\| or it breaks the table:\n%s", got)
	}
}

func TestCVLines(t *testing.T) {
	got := CVLines(load(t, good))
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("CVLines gave %d lines, want 3:\n%s", len(lines), got)
	}
	want := "- Webhook Relay: Built an HMAC-verified webhook relay with replay protection and exponential backoff. (https://github.com/aminaw/relay)"
	if lines[0] != want {
		t.Errorf("first line\n got: %s\nwant: %s", lines[0], want)
	}
}

func TestRenderPage(t *testing.T) {
	var b strings.Builder
	if err := RenderPage(&b, load(t, good)); err != nil {
		t.Fatalf("RenderPage: %v", err)
	}
	html := b.String()
	for _, s := range []string{
		"<!doctype html>", `<html lang="en">`, `name="viewport"`,
		"Amina Wanjiru", "build backends that don&#39;t lose money",
		`href="https://github.com/aminaw/kvstore"`, `href="https://relay.example.dev"`,
		`class="featured"`, "storage",
	} {
		if !strings.Contains(html, s) {
			t.Errorf("page has no %q", s)
		}
	}
	relay, kv := strings.Index(html, "Webhook Relay"), strings.Index(html, "Key-Value Store")
	if relay < 0 || kv < 0 || relay > kv {
		t.Errorf("the featured Webhook Relay should come before Key-Value Store")
	}
	if strings.Count(html, `class="featured"`) != 1 {
		t.Errorf("exactly one card should be featured")
	}
}

func TestRenderPageEscapes(t *testing.T) {
	p := Portfolio{Owner: "<script>alert(1)</script>", Projects: []Project{
		{Name: "<b>Bold</b>", Day: 1, Summary: "x & y", Repo: "https://github.com/a/b", CVLine: "x"},
	}}
	var b strings.Builder
	if err := RenderPage(&b, p); err != nil {
		t.Fatal(err)
	}
	html := b.String()
	if strings.Contains(html, "<script>alert") || strings.Contains(html, "<b>Bold") {
		t.Errorf("RenderPage didn't escape user text; use html/template:\n%s", html)
	}
	if !strings.Contains(html, "x &amp; y") {
		t.Errorf("want the summary escaped as x &amp; y")
	}
}
