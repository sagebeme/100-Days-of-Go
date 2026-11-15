# How we write code in this course

Ten habits for every day, from Day 1 to Day 100. They're what a reviewer at work checks first, so practise them now while the code is small.

## Before you write

1. **Say what "done" means.** Read the brief, the steps and the test names, then say in one line what your function must do. If the brief and a test seem to disagree, or you can't tell what an input should give, stop and ask (or reread the test: it is the spec).
2. **Write only what the day asks for.** No extra flags, options or helpers "for later". If you think of something worth adding, note it as a stretch goal.

## While you write

3. **Keep functions small and clear.** A function much longer than a screen usually splits into two. Before you finish, delete anything you added and didn't use: a variable, a helper, a commented-out try.
4. **Change only what the task needs.** Don't rename, reorder or reformat code the step didn't mention, and don't change a function's name or parameters: the tests call it exactly like that. Match the names and style of the code around you.
5. **Don't hide errors.** Check every `err`. Either handle it, or add context and pass it up with `%w` ([Day 17](../phase-2-modelling/day-017-failed-transactions/)). Never drop one with `_` without a comment saying why it can't fail. Use `panic` only for "this can't happen" cases, with a comment.

```go
// Bad: if the file is missing, data is empty and nobody knows why.
data, _ := os.ReadFile(path)

// Good: the caller learns what failed, and errors.Is still works.
data, err := os.ReadFile(path)
if err != nil {
	return fmt.Errorf("loading budget %s: %w", path, err)
}
```

   Katas are the one exception: the judge promises valid input, so the fast-input template in [big-o.md](big-o.md#fast-input-and-output) skips the `strconv.Atoi` error. Nowhere else.

6. **Comments say why, not what.** A doc comment starts with the name and says what the function takes, what it returns, and when it returns an error: `// Total returns the basket's price in shillings, or an error if an item has no price.` `// loop over items` above a loop adds nothing.

## Bugs and tests

7. **Bug? Write a failing test first.** Add a test (or a table row) that fails because of the bug, watch it fail, then fix the code until it passes, then run all of that day's tests. [Day 97](../phase-8-portfolio/day-097-fix-an-open-source-issue/) is this habit in a real project.
8. **Tests check results.** A test compares what came back with what you expected and calls `t.Errorf` (or `t.Fatalf` when the rest can't run). A test that only prints is a demo: it passes even when the code is wrong. Use a table of cases when you have more than two ([Day 19](../phase-2-modelling/day-019-table-tests/)). A `fmt.Println` to look at a value is fine while you debug; take it out before you're done.

## Before you call it done

9. **Run the checks, from the repo root**, with your day's folder in place of `DAY` (for example `./phase-1-terminal/day-001-mpesa-sms`) and its number in place of `N`:

```bash
gofmt -l DAY/starter        # prints nothing when the code is formatted; gofmt -w fixes it
go vet DAY/starter/...      # prints nothing when it finds nothing
go test DAY/starter         # ok means done; add -race from Day 45 on
go run ./tools/kata N       # the day's kata
```

   In Phase 6, run `npm test -- day-0NN` and `npm run typecheck` from `phase-6-frontend`. Don't run `go test ./...` across the whole repo: untouched starters fail on purpose.

10. **Commit carefully.** Never commit `.env`, keys, tokens, `cover.out`, `dist/` or `node_modules/` ([Day 21](../phase-2-modelling/day-021-config-loader/), [Day 25](../phase-2-modelling/day-025-git-team/)). Stage files by name or with `git add -p`, not `git add .`. A commit message says what changed and why: `fix: round half-cents up in FormatAmount`, not `update` or `wip`.
