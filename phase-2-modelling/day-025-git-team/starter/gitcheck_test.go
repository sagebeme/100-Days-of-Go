package gitteam

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCommit(t *testing.T) {
	tests := []struct {
		msg  string
		want Commit
	}{
		{"feat: add CSV export", Commit{Type: "feat", Description: "add CSV export"}},
		{"fix(ledger): stop negative balances", Commit{Type: "fix", Scope: "ledger", Description: "stop negative balances"}},
		{"feat(api)!: rename /pay to /payments", Commit{Type: "feat", Scope: "api", Breaking: true, Description: "rename /pay to /payments"}},
		{"refactor!: drop Go 1.20 support", Commit{Type: "refactor", Breaking: true, Description: "drop Go 1.20 support"}},
		{"docs: explain .env.example\n\nNew teammates kept missing it.\n", Commit{Type: "docs", Description: "explain .env.example", Body: "New teammates kept missing it."}},
		{"feat: new statement format\n\nReads the 2026 layout.\n\nBREAKING CHANGE: the old layout is no longer read",
			Commit{Type: "feat", Breaking: true, Description: "new statement format", Body: "Reads the 2026 layout.\n\nBREAKING CHANGE: the old layout is no longer read"}},
		{"chore: tidy\n\nBREAKING-CHANGE: config moved", Commit{Type: "chore", Breaking: true, Description: "tidy", Body: "BREAKING-CHANGE: config moved"}},
		{"Feat: Capital letters parse fine", Commit{Type: "Feat", Description: "Capital letters parse fine"}},
		{"fix(): empty scope parses", Commit{Type: "fix", Scope: "", Description: "empty scope parses"}},
		{"fix: ", Commit{Type: "fix", Description: ""}},
	}
	for _, tt := range tests {
		t.Run(strings.SplitN(tt.msg, "\n", 2)[0], func(t *testing.T) {
			got, err := ParseCommit(tt.msg)
			if err != nil || got != tt.want {
				t.Errorf("ParseCommit(%q)\n got: %+v, %v\nwant: %+v, <nil>", tt.msg, got, err, tt.want)
			}
		})
	}
}

func TestParseCommitRejects(t *testing.T) {
	bad := []string{
		"",
		"add CSV export",
		"feat:add CSV export", // no space after the colon
		"feat add: export",    // a space in the type
		": no type",
		"feat(api: broken bracket",
		"fe4t: digits in the type",
		"(api): scope but no type",
	}
	for _, msg := range bad {
		if got, err := ParseCommit(msg); err == nil {
			t.Errorf("ParseCommit(%q) = %+v, nil; want an error", msg, got)
		}
	}
}

func TestLintCommitGood(t *testing.T) {
	good := []string{
		"feat: add CSV export",
		"fix(ledger): stop negative balances",
		"feat(mpesa-api)!: rename /pay to /payments",
		"docs: explain .env.example\n\nNew teammates kept missing it.",
		"test(day-24): cover the byte order mark",
		"chore: " + strings.Repeat("x", 65), // exactly 72 characters
		"fix: handle shillings with no cents\n\nA cell like \"750\" crashed the parser.\n\nCloses #42",
	}
	for _, msg := range good {
		got := LintCommit(msg)
		if got == nil || len(got) != 0 {
			t.Errorf("LintCommit(%q) = %#v, want []string{}", msg, got)
		}
	}
}

func TestLintCommitProblems(t *testing.T) {
	types := "build, chore, ci, docs, feat, fix, perf, refactor, revert, style, test"
	tests := []struct {
		name string
		msg  string
		want []string
	}{
		{"empty", "", []string{"header is empty"}},
		{"blank header", "   \n\nbody", []string{"header is empty"}},
		{"not conventional", "Fixed the thing", []string{"header must look like type(scope): description"}},
		{"no space after colon", "fix:the thing", []string{"header must look like type(scope): description"}},
		{"unknown type", "feature: add export", []string{`unknown type "feature"; use one of ` + types}},
		{"capital type", "Fix: the thing", []string{`unknown type "Fix"; use one of ` + types}},
		{"capital scope", "feat(API): add export", []string{`scope "API" must be lowercase letters, digits and hyphens`}},
		{"empty scope", "feat(): add export", []string{`scope "" must be lowercase letters, digits and hyphens`}},
		{"empty description", "fix: ", []string{"description is empty"}},
		{"two spaces", "fix:  the thing", []string{"description must not start with a space"}},
		{"full stop", "fix: the thing.", []string{"description must not end with a full stop"}},
		{"too long", "feat: " + strings.Repeat("a", 74), []string{"header is 80 characters; keep it to 72 or fewer"}},
		{"no blank line", "fix: the thing\nMore detail here.", []string{"leave a blank line between the header and the body"}},
		{"empty breaking change", "feat!: new format\n\nBREAKING CHANGE:", []string{"BREAKING CHANGE needs a description after it"}},
		{"several at once", "Feature(Ledger): Fixed it all and wrote a long header about it to be sure.\nno blank line",
			[]string{
				`unknown type "Feature"; use one of ` + types,
				`scope "Ledger" must be lowercase letters, digits and hyphens`,
				"description must not end with a full stop",
				"header is 74 characters; keep it to 72 or fewer",
				"leave a blank line between the header and the body",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LintCommit(tt.msg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LintCommit(%q)\n got: %q\nwant: %q", tt.msg, got, tt.want)
			}
		})
	}
}

func TestLintCommitCountsRunes(t *testing.T) {
	// "ü" is two bytes but one character: this header is 72 characters, 73 bytes.
	msg := "docs: " + strings.Repeat("a", 65) + "ü"
	if got := LintCommit(msg); len(got) != 0 {
		t.Errorf("LintCommit of a 72-character header with a two-byte letter = %q, want no problems", got)
	}
}

func TestCheckBranchGood(t *testing.T) {
	for _, name := range []string{
		"feat/mpesa-export",
		"fix/123-negative-balance",
		"docs/readme",
		"chore/bump-pgx-to-v5-11",
		"test/day-24",
		"feat/" + strings.Repeat("a", 55), // exactly 60
	} {
		if err := CheckBranch(name); err != nil {
			t.Errorf("CheckBranch(%q) = %v, want nil", name, err)
		}
	}
}

func TestCheckBranchProblems(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"main", `"main" is protected; make a branch: git switch -c feat/your-change`},
		{"master", `"master" is protected`},
		{"develop", `"develop" is protected`},
		{"feat/" + strings.Repeat("a", 56), "branch name is 61 characters; keep it to 60 or fewer"},
		{"add-export", `branch "add-export" must look like type/short-description`},
		{"feat/", `branch "feat/" must look like type/short-description`},
		{"/export", `branch "/export" must look like type/short-description`},
		{"feature/export", `unknown branch type "feature"; use one of build, chore, ci`},
		{"Feat/export", `unknown branch type "Feat"`},
		{"feat/Mpesa_Export", `"Mpesa_Export" must be lowercase letters, digits and single hyphens`},
		{"feat/mpesa export", `"mpesa export" must be lowercase letters`},
		{"feat/-export", `"-export" must be lowercase letters`},
		{"feat/export-", `"export-" must be lowercase letters`},
		{"feat/mpesa--export", `"mpesa--export" must be lowercase letters`},
		{"feat/mpesa/export", `"mpesa/export" must be lowercase letters`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckBranch(tt.name)
			if err == nil {
				t.Fatalf("CheckBranch(%q) = nil, want an error containing %q", tt.name, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CheckBranch(%q)\n got: %v\nwant it to contain: %s", tt.name, err, tt.want)
			}
		})
	}
}
