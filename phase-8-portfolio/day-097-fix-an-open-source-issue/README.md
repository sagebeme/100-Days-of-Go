# Day 97: Fix an Open-Source Issue

## The brief

`pesa` is a small open-source Go package (made up for this day) that parses and formats Kenyan shilling amounts. Hundreds of apps use it. Its code is in `starter/`, exactly as the maintainers left it, with their `CONTRIBUTING.md`. A user has opened this issue:

> **#42 Parse reads "1,200.5" as Ksh 1,200.05**
>
> *opened by mwangi-dev*
>
> Our chama app lets members type their contribution. A member typed `1,200.5` and the receipt said **Ksh 1,200.05**. Only 45 cents went missing, but now the chama's totals don't match the M-Pesa statement and the treasurer can't close the month.
>
> ```go
> c, _ := pesa.Parse("Ksh 1,200.5")
> fmt.Println(c, pesa.Format(c))
> // got:  120005 Ksh 1,200.05
> // want: 120050 Ksh 1,200.50
> ```
>
> pesa v1.4.0, Go 1.26, Linux. Two decimal places (`1,200.50`) work fine.
>
> *Labels: bug, good first issue*

Your job is to be the contributor who fixes it: reproduce it, find the cause in code you didn't write, make the smallest correct fix, and open a pull request a busy maintainer can merge in five minutes.

**Acceptance criteria**

- A new test that fails on the current code with the input from the issue, and passes after your fix.
- A fix that changes only what the issue needs. Every existing test still passes, and so do the maintainers' tests in `maintainer_test.go`.
- `PR_DESCRIPTION.md` filled in from the template.

## What you'll use

- Branches, small commits and pull requests, from [Day 25: Git Like a Team](../../phase-2-modelling/day-025-git-team/).
- Table-driven tests, from [Day 19](../../phase-2-modelling/day-019-table-tests/).
- Errors that wrap a sentinel, from [Day 17](../../phase-2-modelling/day-017-failed-transactions/), because you'll read a `SyntaxError` with an `Unwrap` method.

### Finding an issue worth your time

- Search GitHub for `is:issue is:open label:"good first issue" language:Go`, or browse [goodfirstissue.dev](https://goodfirstissue.dev) and [up-for-grabs.net](https://up-for-grabs.net). Better still, pick a library you already use: you understand its users.
- Check the project is alive: commits in the last month, and maintainers who answer pull requests. A PR to a dead repo will sit for ever.
- Pick a bug with a clear reproduction. Avoid issues where people are still arguing about what the behaviour should be.
- Read the whole thread. Someone may have tried already, or a maintainer may have said how they want it fixed.

### Before you write code

- Read `CONTRIBUTING.md`. It tells you what CI runs, how commits are named, whether you need an issue first, and sometimes that you must sign a CLA (a contributor licence agreement).
- Comment on the issue: "I'd like to work on this. My plan is X." Wait for a yes on anything bigger than a typo.

### Etiquette

- Small PRs get merged; big ones get "can you split this up?". One fix per PR, with no drive-by reformatting.
- Match the project's style even where you'd do it differently.
- Answer review comments with new commits, not a force-push, unless the project asks for one. Thank people. Don't ping maintainers every day; a polite nudge after a week or two is fine.
- If a maintainer says no, ask what they'd accept, and take the answer gracefully. It's their project.

## Steps

Requirements. Work as if this were a real repository.

**Must-haves**

1. Read `starter/CONTRIBUTING.md`, then run the project's checks before you change anything. Every test in `pesa_test.go` passes; `maintainer_test.go` fails, because it holds the tests the maintainers added when they merged the fix. Don't open it yet.
2. **Reproduce.** Create `starter/issue42_test.go` with a test that uses the exact input from the issue. Run it and see it fail with the issue's numbers. A fix without a failing test first is a guess.
3. **Find the cause.** Read `Parse` from the top. Add a `fmt.Println` or use a debugger if you like, but take it out again. Write down in one sentence why `"5"` becomes 5 cents.
4. **Fix it**, as small as possible: no renames, no new exported names, no changes to `Format` or `Split`. Add a short comment that names the issue.
5. Run what CONTRIBUTING asks for, and the maintainers' tests:

```bash
gofmt -l ./phase-8-portfolio/day-097-fix-an-open-source-issue/starter
go vet ./phase-8-portfolio/day-097-fix-an-open-source-issue/starter
go test -race ./phase-8-portfolio/day-097-fix-an-open-source-issue/starter
```

6. Fill in `starter/PR_DESCRIPTION.md`: what it fixes with `Fixes #42`, the cause, the fix and why this way, how you tested it, and what's out of scope.
7. Write the commit message in the project's style: `pesa: <what changed>`.

**Stretch goals**

- Hunt for another bug. Write tests for edge cases the project doesn't test (`Split` with `n` bigger than the amount, `Format` of the most negative `Cents`, very long inputs). If you find one, write it up as an issue in the style of #42: input, got, want, version.
- Add a fuzz test, `FuzzParseFormat`, checking that `Parse(Format(c)) == c` for any `c` in range. Fuzzing finds the bugs nobody thought to test.
- Do it for real: find a `good first issue` in a Go project you use, and open a pull request. Link it in your portfolio, merged or not.

## The kata: Amount Parser

Read amounts typed in every format and print them in one standard form, or `INVALID`. It's `pesa.Parse` and `pesa.Format` written from a spec, edge cases and all. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 97
```

## Check yourself

1. Why write a failing test with the exact input from the issue *before* touching `Parse`? What does it prove after the fix?
2. Predict what the current (unfixed) `Parse` returns, in cents:

   ```go
   Parse("Ksh 1,200.5")
   Parse("1,200.50")
   Parse("0.5")
   Parse("-3.5")
   ```

   In one sentence: why does `"5"` become 5 cents?
3. Spot the problem. This PR fixes the bug, and the maintainer asks you to split it up:

   ```
   pesa: fix decimals, rename ungroup to stripGroups, run gofumpt, add Format options
    pesa.go        | 214 ++++++++------
    pesa_test.go   |  96 +++---
    format.go      |  58 +++++
   ```

4. After you open the PR, a maintainer asks for a change. Do you push a new commit or force-push? And what goes in `PR_DESCRIPTION.md` besides "Fixes #42"?

<details>
<summary>Answers</summary>

1. A fix without a failing test first is a guess: the test proves you reproduced the bug the user saw. Afterwards it proves the fix works, and it stays as a regression test so the bug can't come back unnoticed.
2. `120005`, `120050`, `5`, `-305`. The digits after the point are parsed as a whole number of cents, so one digit means tenths of a shilling but is read as cents; a one-digit fraction needs multiplying by 10.
3. Four ideas in one PR: a fix, a rename, a reformat and a feature. It's slow to review and risky to merge, and the reformat hides the real change. One fix per PR, in the project's style, with no drive-by changes: the fix and its test, and nothing else.
4. Usually a new commit, unless `CONTRIBUTING.md` asks for a force-push, so the reviewer can see what changed since they last looked. The description says what it fixes, the cause, the fix and why this way, how you tested it, and what's out of scope.

</details>

## When you're stuck

- **Your test passes before the fix**: it doesn't use the issue's input. Use `"Ksh 1,200.5"` exactly, and check the number of cents, not just that there's no error.
- **You can't see where cents come from**: follow `frac` from the `strings.Cut` to the line that adds it to the total. Ask what `strconv.ParseInt("5", 10, 64)` returns and what `.5` means.
- **Your fix breaks `TestParse`**: you changed how two digits are read. Only one digit after the point needs changing.
- **`gofmt -l` prints your file**: run `gofmt -w` on it. CI would fail on that before a human ever reads your PR.
- **Your PR description is a list of what you typed**: describe the behaviour, not the edits. A reviewer wants the why; they can read the diff for the what.
- **Still stuck?** Open `solution/`, read `pesa.go`, `issue42_test.go` and `PR_DESCRIPTION.md` once, close them, then do your own version from memory.

## Make it portfolio-ready

- **README**: a "Open source" section in your portfolio listing every PR you've opened, with one line each: the project, the problem, the fix, merged or not. Reviewers read merged PRs as proof you can work in someone else's codebase.
- **Demo**: link the PR itself. The issue, your test, the review comments and your answers are the demo.
- **Tests and CI**: your PR shows the project's CI passing. Mention it, and link the test you added.
- **Deploy link**: the release that shipped your fix, for example "fixed in v1.4.1".
- **What you'd do next**: the follow-up issue you opened, or the next bug you'd like to take on in the same project. Staying with one project builds trust, and trust gets you bigger work.
