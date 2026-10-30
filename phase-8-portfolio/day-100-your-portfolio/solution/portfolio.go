// Package portfolio turns a projects.json file into a portfolio page and a README
// table of CI badges, so your list of projects lives in one place.
package portfolio

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"
)

// Project is one thing you built.
type Project struct {
	Name     string   `json:"name"`
	Day      int      `json:"day"`                // the course day it came from, 1 to 100
	Summary  string   `json:"summary"`            // one sentence, at most MaxSummary characters
	Repo     string   `json:"repo"`               // https://github.com/<owner>/<repo>
	Demo     string   `json:"demo,omitempty"`     // an https link, or empty
	Tags     []string `json:"tags,omitempty"`     // e.g. "go", "postgres"
	Featured bool     `json:"featured,omitempty"` // shown first, with more space
	CVLine   string   `json:"cv_line"`            // one line for a CV or LinkedIn
}

// Portfolio is the whole file.
type Portfolio struct {
	Owner    string    `json:"owner"`
	Headline string    `json:"headline"`
	Projects []Project `json:"projects"`
}

// MaxSummary is the longest summary allowed, in characters.
const MaxSummary = 160

// Workflow is the GitHub Actions workflow file the badges point at.
const Workflow = "ci.yml"

// Load reads a portfolio from JSON, rejecting unknown fields, and validates it.
func Load(r io.Reader) (Portfolio, error) {
	var p Portfolio
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Portfolio{}, fmt.Errorf("portfolio: reading JSON: %w", err)
	}
	if err := Validate(p); err != nil {
		return Portfolio{}, err
	}
	return p, nil
}

// Validate checks the whole portfolio and returns every problem at once, joined with
// errors.Join. Each message names the project and the field.
func Validate(p Portfolio) error {
	var errs []error
	if strings.TrimSpace(p.Owner) == "" {
		errs = append(errs, errors.New("owner: required"))
	}
	if len(p.Projects) == 0 {
		errs = append(errs, errors.New("projects: add at least one"))
	}
	seen := map[string]bool{}
	for i, pr := range p.Projects {
		label := fmt.Sprintf("project %d (%q)", i+1, pr.Name)
		bad := func(field, msg string) { errs = append(errs, fmt.Errorf("%s: %s: %s", label, field, msg)) }
		name := strings.TrimSpace(pr.Name)
		switch {
		case name == "":
			bad("name", "required")
		case seen[strings.ToLower(name)]:
			bad("name", "used twice")
		}
		seen[strings.ToLower(name)] = true
		if pr.Day < 1 || pr.Day > 100 {
			bad("day", "must be 1 to 100")
		}
		if n := utf8.RuneCountInString(pr.Summary); n == 0 || n > MaxSummary {
			bad("summary", fmt.Sprintf("must be 1 to %d characters", MaxSummary))
		}
		if _, _, err := RepoPath(pr.Repo); err != nil {
			bad("repo", err.Error())
		}
		if pr.Demo != "" && !isHTTPS(pr.Demo) {
			bad("demo", "must be an https:// link")
		}
		if strings.TrimSpace(pr.CVLine) == "" {
			bad("cv_line", "required")
		}
	}
	return errors.Join(errs...)
}

// RepoPath returns the owner and repository from a GitHub URL of the exact form
// https://github.com/<owner>/<repo>, with nothing after it.
func RepoPath(repo string) (owner, name string, err error) {
	u, err := url.Parse(repo)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", errors.New("must look like https://github.com/<owner>/<repo>")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.HasSuffix(parts[1], ".git") {
		return "", "", errors.New("must look like https://github.com/<owner>/<repo>")
	}
	return parts[0], parts[1], nil
}

func isHTTPS(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

// Sorted returns the projects with featured ones first, then the latest course day
// first, then by name. It doesn't change its argument.
func Sorted(ps []Project) []Project {
	out := slices.Clone(ps)
	slices.SortStableFunc(out, func(a, b Project) int {
		if a.Featured != b.Featured {
			if a.Featured {
				return -1
			}
			return 1
		}
		return cmp.Or(cmp.Compare(b.Day, a.Day), strings.Compare(a.Name, b.Name))
	})
	return out
}

// BadgeURL and RunsURL point at the project's CI workflow on GitHub Actions.
func BadgeURL(owner, repo string) string {
	return "https://github.com/" + owner + "/" + repo + "/actions/workflows/" + Workflow + "/badge.svg"
}

// RunsURL is the page the badge links to.
func RunsURL(owner, repo string) string {
	return "https://github.com/" + owner + "/" + repo + "/actions/workflows/" + Workflow
}

// BadgeTable returns a Markdown table, in Sorted order, with a row per project:
//
//	| Project | Day | CI | Demo |
//	|---|---|---|---|
//	| [Webhook Relay](https://github.com/amina/relay) | 95 | [![CI](<badge>)](<runs>) | [demo](https://...) |
//
// A project without a demo shows "-". A "|" in a name is escaped as "\|".
func BadgeTable(p Portfolio) string {
	var b strings.Builder
	b.WriteString("| Project | Day | CI | Demo |\n|---|---|---|---|\n")
	for _, pr := range Sorted(p.Projects) {
		owner, repo, _ := RepoPath(pr.Repo)
		demo := "-"
		if pr.Demo != "" {
			demo = "[demo](" + pr.Demo + ")"
		}
		fmt.Fprintf(&b, "| [%s](%s) | %d | [![CI](%s)](%s) | %s |\n",
			strings.ReplaceAll(pr.Name, "|", `\|`), pr.Repo, pr.Day, BadgeURL(owner, repo), RunsURL(owner, repo), demo)
	}
	return b.String()
}

// CVLines returns one line per project, in Sorted order: "- <Name>: <CVLine> (<Repo>)".
func CVLines(p Portfolio) string {
	var b strings.Builder
	for _, pr := range Sorted(p.Projects) {
		fmt.Fprintf(&b, "- %s: %s (%s)\n", pr.Name, pr.CVLine, pr.Repo)
	}
	return b.String()
}

// RenderPage writes the portfolio as one HTML page: the owner and headline, then a
// card per project in Sorted order with its name, summary, tags, a link to the code
// and, if there is one, the demo. Featured cards have the class "featured".
func RenderPage(w io.Writer, p Portfolio) error {
	data := struct {
		Portfolio
		Sorted []Project
	}{p, Sorted(p.Projects)}
	var b strings.Builder
	if err := page.Execute(&b, data); err != nil {
		return err
	}
	_, err := io.WriteString(w, b.String())
	return err
}

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Owner}}: projects</title>
<style>
body { font: 16px/1.5 system-ui, sans-serif; max-width: 46rem; margin: 2rem auto; padding: 0 1rem; color: #1a1a1a; background: #fff; }
article { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; margin: 1rem 0; }
article.featured { border: 2px solid #0a7d4b; }
.tags { color: #555; font-size: 0.9rem; }
</style>
</head>
<body>
<header>
<h1>{{.Owner}}</h1>
<p>{{.Headline}}</p>
</header>
<main>
{{range .Sorted}}<article{{if .Featured}} class="featured"{{end}}>
<h2>{{.Name}}</h2>
<p>{{.Summary}}</p>
{{with .Tags}}<p class="tags">{{range $i, $t := .}}{{if $i}} · {{end}}{{$t}}{{end}}</p>{{end}}
<p><a href="{{.Repo}}">Code</a>{{with .Demo}} · <a href="{{.}}">Live demo</a>{{end}}</p>
</article>
{{end}}</main>
</body>
</html>
`))
