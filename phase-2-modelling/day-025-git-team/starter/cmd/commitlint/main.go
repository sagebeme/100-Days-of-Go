// Command commitlint checks a commit message file. Git runs it as a commit-msg hook
// with the path of the message as its only argument; a non-zero exit stops the commit.
//
//	go run ./phase-2-modelling/day-025-git-team/starter/cmd/commitlint .git/COMMIT_EDITMSG
package main

import (
	"fmt"
	"os"
	"strings"

	gitteam "github.com/sagebeme/100-Days-of-Go/phase-2-modelling/day-025-git-team/starter"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: commitlint <commit-message-file>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "commitlint:", err)
		os.Exit(2)
	}

	// Git's editor template adds lines starting with # that never reach the commit.
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	msg := strings.TrimRight(strings.Join(lines, "\n"), "\n")

	problems := gitteam.LintCommit(msg)
	if len(problems) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "commit message problems:")
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "  -", p)
	}
	os.Exit(1)
}
