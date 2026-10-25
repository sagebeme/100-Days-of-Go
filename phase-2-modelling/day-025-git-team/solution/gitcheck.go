// Package gitteam checks the two things a team's git hooks check most: that a commit
// message follows Conventional Commits, and that a branch name follows the team's
// pattern.
package gitteam

import (
	"fmt"
	"slices"
	"strings"
)

// Types are the commit types the team allows, in alphabetical order.
var Types = []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}

// MaxHeader is the longest a commit header may be.
const MaxHeader = 72

// Commit is a parsed commit message.
type Commit struct {
	Type        string // "feat"
	Scope       string // "ledger" in "feat(ledger): ..."; "" if none
	Breaking    bool   // "!" before the colon, or a BREAKING CHANGE footer
	Description string // the text after ": "
	Body        string // everything after the blank line, trimmed; "" if none
}

// ParseCommit splits a commit message into its parts. The first line is the header:
//
//	type(scope)!: description
//
// The scope in brackets and the "!" are optional. It returns an error if the header
// has no ": ", the type is empty or isn't made of letters, or the brackets are
// broken. ParseCommit doesn't judge the parts; LintCommit does.
//
// Breaking is true if the header has "!" or any line of the body starts with
// "BREAKING CHANGE: " or "BREAKING-CHANGE: ".
func ParseCommit(msg string) (Commit, error) {
	header, rest, _ := strings.Cut(msg, "\n")
	prefix, desc, ok := strings.Cut(header, ": ")
	if !ok {
		return Commit{}, fmt.Errorf("header %q must look like type(scope): description", header)
	}
	var c Commit
	c.Description = desc
	if p, found := strings.CutSuffix(prefix, "!"); found {
		c.Breaking = true
		prefix = p
	}
	if open := strings.IndexByte(prefix, '('); open != -1 {
		if !strings.HasSuffix(prefix, ")") {
			return Commit{}, fmt.Errorf("header %q: the scope needs a closing bracket", header)
		}
		c.Scope = prefix[open+1 : len(prefix)-1]
		prefix = prefix[:open]
	}
	if prefix == "" || !letters(prefix) {
		return Commit{}, fmt.Errorf("header %q must start with a type made of letters", header)
	}
	c.Type = prefix
	c.Body = strings.TrimSpace(rest)
	for _, line := range strings.Split(c.Body, "\n") {
		if strings.HasPrefix(line, "BREAKING CHANGE: ") || strings.HasPrefix(line, "BREAKING-CHANGE: ") {
			c.Breaking = true
		}
	}
	return c, nil
}

func letters(s string) bool {
	for i := 0; i < len(s); i++ {
		if !(s[i] >= 'a' && s[i] <= 'z' || s[i] >= 'A' && s[i] <= 'Z') {
			return false
		}
	}
	return true
}

// slug reports whether s is lowercase letters, digits and single hyphens, not starting
// or ending with a hyphen.
func slug(s string) bool {
	if s == "" || s[0] == '-' || s[len(s)-1] == '-' || strings.Contains(s, "--") {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

// LintCommit returns every problem with a commit message, in this order, or an empty
// (non-nil) slice if there are none:
//
//	header is empty                                    (and nothing else is checked)
//	header must look like type(scope): description     (ParseCommit failed; nothing else is checked)
//	unknown type "Feat"; use one of build, chore, ci, docs, feat, fix, perf, refactor, revert, style, test
//	scope "API" must be lowercase letters, digits and hyphens
//	description is empty
//	description must not start with a space
//	description must not end with a full stop
//	header is 80 characters; keep it to 72 or fewer
//	leave a blank line between the header and the body
//	BREAKING CHANGE needs a description after it
//
// The scope check only applies when there are brackets. The last check looks for a
// body line that is exactly "BREAKING CHANGE:" or "BREAKING-CHANGE:" with nothing after
// it. Header length counts characters (runes), not bytes.
func LintCommit(msg string) []string {
	problems := []string{}
	header, rest, hasRest := strings.Cut(msg, "\n")
	if strings.TrimSpace(header) == "" {
		return append(problems, "header is empty")
	}
	c, err := ParseCommit(msg)
	if err != nil {
		return append(problems, "header must look like type(scope): description")
	}

	if !slices.Contains(Types, c.Type) {
		problems = append(problems, fmt.Sprintf("unknown type %q; use one of %s", c.Type, strings.Join(Types, ", ")))
	}
	if strings.Contains(header, "(") && !slug(c.Scope) {
		problems = append(problems, fmt.Sprintf("scope %q must be lowercase letters, digits and hyphens", c.Scope))
	}
	switch {
	case strings.TrimSpace(c.Description) == "":
		problems = append(problems, "description is empty")
	case strings.HasPrefix(c.Description, " "):
		problems = append(problems, "description must not start with a space")
	}
	if strings.HasSuffix(c.Description, ".") {
		problems = append(problems, "description must not end with a full stop")
	}
	if n := len([]rune(header)); n > MaxHeader {
		problems = append(problems, fmt.Sprintf("header is %d characters; keep it to %d or fewer", n, MaxHeader))
	}
	if hasRest {
		second, _, _ := strings.Cut(rest, "\n")
		if strings.TrimSpace(second) != "" {
			problems = append(problems, "leave a blank line between the header and the body")
		}
	}
	for _, line := range strings.Split(rest, "\n") {
		if t := strings.TrimSpace(line); t == "BREAKING CHANGE:" || t == "BREAKING-CHANGE:" {
			problems = append(problems, "BREAKING CHANGE needs a description after it")
			break
		}
	}
	return problems
}

// Protected are branches nobody commits to directly.
var Protected = []string{"main", "master", "develop"}

// MaxBranch is the longest a branch name may be.
const MaxBranch = 60

// CheckBranch returns the first problem with a branch name, or nil. A good name is a
// type from Types, a slash, then a short description in lowercase letters, digits and
// single hyphens, like "feat/mpesa-export" or "fix/123-negative-balance". It checks,
// in order:
//
//	"main" is protected; make a branch: git switch -c feat/your-change
//	branch name is 61 characters; keep it to 60 or fewer
//	branch "add-export" must look like type/short-description
//	unknown branch type "feature"; use one of build, chore, ...
//	"Mpesa_Export" must be lowercase letters, digits and single hyphens
func CheckBranch(name string) error {
	if slices.Contains(Protected, name) {
		return fmt.Errorf("%q is protected; make a branch: git switch -c feat/your-change", name)
	}
	if n := len([]rune(name)); n > MaxBranch {
		return fmt.Errorf("branch name is %d characters; keep it to %d or fewer", n, MaxBranch)
	}
	typ, desc, ok := strings.Cut(name, "/")
	if !ok || typ == "" || desc == "" {
		return fmt.Errorf("branch %q must look like type/short-description", name)
	}
	if !slices.Contains(Types, typ) {
		return fmt.Errorf("unknown branch type %q; use one of %s", typ, strings.Join(Types, ", "))
	}
	if !slug(desc) {
		return fmt.Errorf("%q must be lowercase letters, digits and single hyphens", desc)
	}
	return nil
}
