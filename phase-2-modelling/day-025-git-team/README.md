# Day 25: Git Like a Team

## The brief

You've joined a team of four building a payments service in Westlands. On day one the lead says: "Branch off `main`, keep commits small, write messages in Conventional Commits, open a PR, get one approval. CI checks the messages and the branch names." Then your first PR comes back with eleven comments, and four of them are about git, not code.

Today you learn the team habits, and you write the checks a team's CI and git hooks run: a commit message linter and a branch name checker.

```
LintCommit("feat(ledger): add CSV export")     →  []                     (fine)
LintCommit("Fixed stuff.")                     →  ["header must look like type(scope): description"]
LintCommit("feat(API): Add export.")           →  [`scope "API" must be lowercase letters, digits and hyphens`,
                                                   "description must not end with a full stop"]
ParseCommit("feat(api)!: rename /pay to /payments")
  →  Commit{Type: "feat", Scope: "api", Breaking: true, Description: "rename /pay to /payments"}, nil

CheckBranch("fix/123-negative-balance")        →  nil
CheckBranch("main")                            →  "main" is protected; make a branch: git switch -c feat/your-change
CheckBranch("feature/Mpesa_Export")            →  unknown branch type "feature"; use one of build, chore, ...
```

## What you'll use

### Branches and small commits

Never commit straight to `main`. `main` is what's deployed, or about to be. Every change starts on its own branch:

```bash
git switch main && git pull           # start from the latest main
git switch -c feat/mpesa-export       # a new branch, named for the change
# ...edit...
git add -p                            # stage piece by piece, reading each change
git commit                            # one idea per commit
git push -u origin feat/mpesa-export  # share it and open a pull request
```

**One idea per commit.** "Rename a function" and "fix the rounding bug" are two commits, even when you did them in the same hour. Small commits are easy to review, easy to revert, and `git bisect` can find the one that broke something. `git add -p` lets you split a messy working tree into clean commits.

### Conventional Commits

```
type(scope)!: description

Optional body: why the change was made, not what (the diff shows what).

Optional footers, like BREAKING CHANGE: what breaks and how to update.
```

- `feat` adds a feature, `fix` fixes a bug, and `docs`, `test`, `refactor`, `perf`, `chore`, `ci`, `build`, `style` and `revert` cover the rest.
- The scope says which part of the code changed: `feat(ledger): ...`.
- `!` or a `BREAKING CHANGE:` footer warns that the change breaks callers. On Day 20 that meant a major version; today's kata turns commits into the next version.
- Keep the header to 72 characters, so it isn't cut off in `git log --oneline` and on GitHub. Write the description as a command (`add CSV export`, not `added` or `adds`) with no full stop.

### Pull requests and code review

A pull request (PR) asks the team to merge your branch. A good one:

- **is small**: a few hundred lines at most. A 2,000-line PR gets a skim and an "LGTM", which helps no one.
- **says why** in the description, how you tested it, and anything the reviewer should look at first.
- **is green** before you ask for review: tests, `go vet` and `gofmt` all pass in CI.

When you review, review the code, not the person. Ask questions ("What happens if the CSV is empty?") instead of giving orders. Mark small things as optional ("nit: ..."). Approve when it's good enough to merge, not when it's how you'd have written it. When you're reviewed, answer every comment, push fixes as new commits so the reviewer can see what changed, and don't take it personally.

### Merging and rebasing

While you work, `main` moves on. There are two ways to catch up:

| | `git merge origin/main` | `git rebase origin/main` |
|---|---|---|
| What it does | adds a merge commit that joins the two histories | replays your commits on top of the new `main` |
| History | true, but with extra merge commits | a straight line, as if you started today |
| Rewrites commits? | no | yes: new commit IDs |
| Safe on a shared branch? | yes | no |

The rule: **rebase your own branch only, before it's merged, and never rebase `main` or a branch someone else is working on.** After a rebase, push with `git push --force-with-lease`, which refuses if someone else pushed in the meantime. Plain `--force` would overwrite their work.

Teams pick one merge style for PRs on GitHub. **Squash and merge** turns the whole PR into one commit on `main`, so the PR title becomes that commit's message, and it must follow Conventional Commits too.

### Never commit secrets

Before your first commit in any repo, check `.gitignore`. This repo's ignores `.env`, build output (`bin/`), coverage files and editor folders. `git check-ignore -v <file>` tells you which rule ignores a file. If you committed something by mistake, `git rm --cached <file>` stops tracking it, but it stays in the history. For a secret that means **rotate it at once**. [reference/env-and-config.md](../../reference/env-and-config.md) explains what to do if a secret leaks.

Read `git diff --cached` before every commit. It takes ten seconds and it's how you catch the API key, the `fmt.Println("DEBUG")` and the 40 MB CSV.

## Why it matters at work

Git habits decide how fast a team can move. Small branches merge without conflicts. Clear commit messages turn `git log` into documentation, and they make automatic changelogs and version numbers possible. A reviewer who can read a PR in ten minutes will review it today, and one who needs two hours will review it next week. Linting messages and branch names in CI sounds fussy until you've tried to find the commit that broke payments in a history of `fix`, `fix again` and `wip`. And one committed secret can cost more than everything else on this page: GitHub's secret scanning and push protection exist because it happens every day.

## Steps

1. Open `starter/gitcheck.go`. Write `ParseCommit` first. Take the header with `strings.Cut(msg, "\n")`, split it at the first `": "`, take a trailing `!` off the prefix with `strings.CutSuffix`, then pull out a `(scope)`. Whatever's left is the type.
2. Look for `BREAKING CHANGE: ` and `BREAKING-CHANGE: ` at the start of a body line.
3. Write `LintCommit`. Start with `problems := []string{}`. Return early for an empty header and for a header `ParseCommit` rejects. Then check each rule in the order the doc comment lists them, appending a message for each one broken. `slices.Contains(Types, c.Type)` checks the type. Count the header with `len([]rune(header))`, since `len` counts bytes.
4. Write `CheckBranch`. Each rule returns its error straight away, in order. A helper that checks "lowercase letters, digits and single hyphens" serves both the scope check and the branch description.
5. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-025-git-team/starter
```

6. Use it as a real git hook in your own copy of the course. Git runs `.git/hooks/commit-msg` with the message file's path, from the top of the repo, and cancels the commit if it exits non-zero:

```bash
cat > .git/hooks/commit-msg <<'HOOK'
#!/bin/sh
exec go run ./phase-2-modelling/day-025-git-team/starter/cmd/commitlint "$1"
HOOK
chmod +x .git/hooks/commit-msg
git commit --allow-empty -m "Fixed stuff."   # refused, with the problems listed
git commit --allow-empty -m "test: try the commit-msg hook"
```

Hooks live in `.git/`, so they aren't committed or shared. Teams share them through a tool or a `make` target, and run the same check in CI, where nobody can skip it. Delete the hook with `rm .git/hooks/commit-msg` when you're done, or keep it.

7. Practise the team flow on your own fork: make a branch with a name `CheckBranch` accepts, make two small commits, push, open a PR on GitHub, read your own diff there as a reviewer would, and squash-merge it.

## The kata: Next Version

Read the commit headers since the last release and work out the next version: a major release for `!`, minor for `feat`, patch for `fix` or `perf`. It's Day 20's semver and today's Conventional Commits working together, the way release tools do it. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 25
```

## When you're stuck

- **`"feat(api)!: x"` gives the type `feat(api)!`**: cut off the `!` first, then the `(scope)`.
- **`TestLintCommitCountsRunes` fails**: `len("ü")` is 2. Count `[]rune(header)`.
- **The empty-header case also reports "header must look like..."**: return straight after `header is empty`.
- **`LintCommit` returns `nil` for a good message**: start with `[]string{}`, not `var problems []string`. The tests tell the two apart.
- **`feat/mpesa--export` passes**: check for `--` and for a hyphen at either end, not only the allowed characters.
- **The hook says `go: command not found`**: git ran it without your shell's `PATH`. Use the full path to `go` in the hook, from `which go`.
- **`error: failed to push some refs` after a rebase**: the branch's history changed, as expected. Push with `git push --force-with-lease`, and only to your own branch.
- **You committed `.env`**: `git rm --cached .env`, commit, and **rotate every secret in it** before anything else. It's in the history now.
- **Still stuck?** Open `solution/gitcheck.go`, read it once, close it, then write your own version from memory.
