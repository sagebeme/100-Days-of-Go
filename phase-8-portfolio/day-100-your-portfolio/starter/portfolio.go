// Package portfolio turns a projects.json file into a portfolio page and a README
// table of CI badges, so your list of projects lives in one place.
package portfolio

import "io"

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
	// Your code here
	return Portfolio{}, nil
}

// Validate checks the whole portfolio and returns every problem at once, joined with
// errors.Join. Each message names the project and the field (as in the JSON):
//
//	owner     required
//	projects  at least one
//	name      required, and no two the same, ignoring case
//	day       1 to 100
//	summary   1 to MaxSummary characters (runes)
//	repo      exactly https://github.com/<owner>/<repo> (see RepoPath)
//	demo      empty, or an https:// link
//	cv_line   required
func Validate(p Portfolio) error {
	// Your code here
	return nil
}

// RepoPath returns the owner and repository from a GitHub URL of the exact form
// https://github.com/<owner>/<repo>, with nothing after it and no ".git".
func RepoPath(repo string) (owner, name string, err error) {
	// Your code here
	return "", "", nil
}

// Sorted returns the projects with featured ones first, then the latest course day
// first, then by name. It doesn't change its argument.
func Sorted(ps []Project) []Project {
	// Your code here
	return nil
}

// BadgeTable returns a Markdown table, in Sorted order, with a row per project:
//
//	| Project | Day | CI | Demo |
//	|---|---|---|---|
//	| [Webhook Relay](https://github.com/amina/relay) | 95 | [![CI](<badge>)](<runs>) | [demo](https://...) |
//
// <runs> is https://github.com/<owner>/<repo>/actions/workflows/ci.yml and <badge> is
// that followed by /badge.svg. A project without a demo shows "-". A "|" in a name is
// escaped as "\|". Every line ends in "\n".
func BadgeTable(p Portfolio) string {
	// Your code here
	return ""
}

// CVLines returns one line per project, in Sorted order: "- <Name>: <CVLine> (<Repo>)",
// each ending in "\n".
func CVLines(p Portfolio) string {
	// Your code here
	return ""
}

// RenderPage writes the portfolio as one HTML page with html/template: a doctype, a
// lang attribute and a viewport meta tag; the owner and headline; then a card per
// project in Sorted order with its name, summary, tags, a link to the code and, if
// there is one, the demo. Featured cards have class="featured".
func RenderPage(w io.Writer, p Portfolio) error {
	// Your code here
	return nil
}
