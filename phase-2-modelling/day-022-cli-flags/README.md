# Day 22: CLI Flags

## The brief

The treasurer of Umoja Women's Group does two sums every month. She shares a bill between members, and she works out loan repayments at the chama's flat rate. Build her a command-line tool, `chama`, with a subcommand for each sum, proper flags, helpful errors and the exit codes scripts expect.

```
$ chama split -total 12005 -members 5
member 1: Ksh2,401
member 2: Ksh2,401
...

$ chama loan -amount 20000
interest: Ksh6,000
total: Ksh26,000
monthly: Ksh8,667

$ chama loan -amount 20000 -rate 101
chama loan: -rate must be from 0 to 100, got 101        (exit code 2)

$ chama split -total 3 -members 5
chama split: Ksh3 can't be split between 5 members      (exit code 1)
```

The whole program is one function, so the tests can run it:

```go
Run([]string{"split", "-total", "10", "-members", "3"}, &stdout, &stderr)
  →  0, stdout "member 1: Ksh4\nmember 2: Ksh3\nmember 3: Ksh3\n"
```

## What you'll use

New today: the **`flag`** package.

- `flag.NewFlagSet("split", flag.ContinueOnError)` makes a set of flags for one subcommand. `ContinueOnError` means `Parse` returns an error instead of exiting the program, which a test needs.
- `total := fs.Int("total", 0, "the amount to share")` defines a flag and returns a **pointer** to its value. Read it with `*total` after parsing.
- `fs.Parse(args)` reads the flags. It accepts `-total 10`, `-total=10`, `--total=10`, and flags in any order. It stops at the first word that isn't a flag; `fs.Args()` and `fs.NArg()` give you what's left.
- `fs.SetOutput(stderr)` sends the flag package's own messages (`flag provided but not defined: -x`, the `-h` help) to your writer.
- `-h` and `-help` make `Parse` return `flag.ErrHelp` after printing the flags. Treat that as success.
- `fs.Visit(func(f *flag.Flag) { ... })` calls the function for each flag that was **set on the command line**. That's how you tell "the user typed `-total 0`" from "the user forgot `-total`".
- **`io.Writer`**: `Run` writes to `stdout` and `stderr` writers instead of `os.Stdout`. The real `main` passes `os.Stdout` and `os.Stderr`; tests pass a `bytes.Buffer` and read what was printed.
- **Exit codes**: `os.Exit(code)` ends the program with a number the shell can read with `echo $?`. Zero means success. Here 1 means "the command line was fine, but it couldn't be done" and 2 means "the command line was wrong", as the `flag` package and many Unix tools do.

```go
func main() {
	os.Exit(cliflags.Run(os.Args[1:], os.Stdout, os.Stderr))
}
```

`main` stays one line, so there's nothing in it to test.

## Why it matters at work

Command-line tools get run by scripts, cron jobs, CI and Dockerfiles, not just people. A reviewer will check that errors go to stderr and results to stdout, so `chama split ... > shares.txt` captures only the shares. They'll check that a failure gives a non-zero exit code, so `set -e` in a CI script stops on it, and that the message says which flag is wrong and why. Writing `Run(args, stdout, stderr) int` instead of calling `os.Exit` and `fmt.Println` all over the place is what makes all of this testable. You'll see the same shape in the `go` command itself and in most serious Go CLIs.

## Steps

1. Open `starter/chama.go`. `Usage`, the exit codes and a `ksh` formatting helper are written. `Split`, `Loan` and `Run` are empty.
2. Write `Split` and `Loan`, and run their tests with `-run 'TestSplit|TestLoan'`. In `Loan`, rounding up `total / months` is `(total + months - 1) / months`.
3. Start `Run` with a `switch args[0]`. Handle no arguments, `version`, `help` and an unknown command first.
4. Write `runSplit(args, stdout, stderr) int`. Make the `FlagSet`, define `-total` and `-members`, `SetOutput(stderr)`, and `Parse`. If `Parse` returns `flag.ErrHelp`, return `ExitOK`; any other error means the flag package has already printed the problem, so return `ExitUsage`. Use `errors.Is(err, flag.ErrHelp)` to tell them apart.
5. Check for leftover arguments with `fs.NArg()`. Check the required flags with `fs.Visit`, then that they're at least 1. Then call `Split` and print.
6. Write `runLoan` the same way, with defaults of 10 for `-rate` and 3 for `-months`, and range checks for both. When you notice you're writing the same checks twice, move them into small helpers.
7. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-022-cli-flags/starter
```

8. Try the real thing:

```bash
go run ./phase-2-modelling/day-022-cli-flags/starter/cmd/chama loan -amount 50000 -months 6
go run ./phase-2-modelling/day-022-cli-flags/starter/cmd/chama loan -h
go run ./phase-2-modelling/day-022-cli-flags/starter/cmd/chama split -total 3 -members 5; echo "exit code $?"
```

`go run` prints its own `exit status 1` line, and its own exit code is always 1 when the program fails. To see your real exit codes, build it first: `go build -o bin/chama ./phase-2-modelling/day-022-cli-flags/starter/cmd/chama && ./bin/chama split -total 3 -members 5; echo $?`. The `bin/` folder is in `.gitignore`, so the binary won't be committed.

## The kata: Flag Parser

Write the core of a flag parser yourself: `-name value`, `-name=value`, bools, `--`, and the exact error messages for each kind of mistake. After it you'll know precisely what `fs.Parse` does. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 22
```

## When you're stuck

- **The whole test run stops with `exit status 2`**: you used `flag.ExitOnError`, or called `os.Exit` inside `Run`. Use `flag.ContinueOnError` and return a code.
- **Flags print to your terminal, not into the test's buffer**: call `fs.SetOutput(stderr)` before `fs.Parse`.
- **Your flags are always 0**: you read the pointer before parsing, or printed `total` (an address) instead of `*total`.
- **`-total 0` and no `-total` behave the same**: use `fs.Visit` to see which flags were actually set.
- **`chama split now -total 5` complains that `-total` is required**: parsing stops at `now`, the first non-flag word, so `-total` is never seen. That's how the `flag` package works; flags go before arguments.
- **`split -h` exits with 2**: check for `flag.ErrHelp` with `errors.Is` and return `ExitOK`.
- **The usage error test sees text on stdout**: errors and help go to `stderr`. Only results go to `stdout`.
- **Still stuck?** Open `solution/chama.go`, read it once, close it, then write your own version from memory.
