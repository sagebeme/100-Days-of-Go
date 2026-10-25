// Package gitteam checks the two things a team's git hooks check most: that a commit
// message follows Conventional Commits, and that a branch name follows the team's
// pattern.
package gitteam

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
	// Your code here.
	return Commit{}, nil
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
	// Your code here.
	return nil
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
	// Your code here.
	return nil
}
