# Phase 2 recap: Modelling and engineering habits

Days 15 to 26. You stopped writing scripts and started writing software: data with a shape, errors a caller can act on, config from the environment, and changes that arrive through pull requests.

## What the phase covered

| Days | Topic |
|---|---|
| 15 to 16 | structs and methods, pointer and value receivers; interfaces and implicit satisfaction |
| 17 to 18 | custom errors, wrapping, `errors.Is` and `errors.As`; generics with type parameters and constraints |
| 19 to 20 | table-driven tests, subtests, helpers and coverage; `go.mod`, versions and dependencies |
| 21 to 22 | env variables, 12-factor config, `.env` files and never committing secrets; the `flag` package, subcommands and exit codes |
| 23 to 24 | files, `encoding/json` and struct tags; `encoding/csv` on a real M-Pesa statement |
| 25 | branches, small commits, pull requests and code review |
| 26 | the capstone: Pesa Budget CLI, configured by env and flags |

## Key ideas

1. **A method with a value receiver gets a copy.** If it must change the struct, use a pointer receiver.
2. **Interfaces are satisfied, not declared.** Write small interfaces where they're used (`Payer`, `Store`), and any type with those methods fits, including test fakes.
3. **Wrap errors with `%w`, check them with `errors.Is` and `errors.As`.** The message gets context ("pay 500: low balance") and the caller can still tell what went wrong.
4. **Generics remove copy-paste, not thinking.** Reach for a type parameter when the same code works for many types; keep plain types otherwise.
5. **A table of cases is a test you can extend in one line.** Name each case, run it with `t.Run`, and the failure tells you which one broke.
6. **Config comes from the environment, checked at startup.** Flags beat env, env beats defaults, and every problem is reported at once. Secrets live in `.env`, which is in `.gitignore`.
7. **Exit codes are part of a CLI's interface.** 0 for success, non-zero for failure, messages on stderr, results on stdout.
8. **Small commits on a branch, reviewed in a PR.** One idea per commit makes history readable and review quick.

## Self-test

1. Predict the output:

   ```go
   type Shop struct{ Name string }
   func (s Shop) Rename(n string)   { s.Name = n }
   func (s *Shop) Rename2(n string) { s.Name = n }

   sh := Shop{"Duka"}
   sh.Rename("X")
   fmt.Println(sh.Name)
   sh.Rename2("Y")
   fmt.Println(sh.Name)
   ```

2. `var ErrLow = errors.New("low balance")`. What does `errors.Is` say for `fmt.Errorf("pay: %w", ErrLow)` and for `fmt.Errorf("pay: %v", ErrLow)`?
3. A user runs `budget -month 2026-10` with `PESA_MONTH=2026-09` set. Which month wins, and why?
4. Why is `json:"amount_ksh"` on a struct field useful, and what happens to an unexported field when you `json.Marshal` the struct?
5. You've committed a `.env` with a real key and pushed it. What do you do first?

<details>
<summary>Answers</summary>

1. `Duka`, then `Y`. `Rename` changes a copy; `Rename2` has a pointer receiver, so it changes `sh` itself.
2. `true` for `%w`, which wraps; `false` for `%v`, which only copies the text.
3. `2026-10`: a flag the user typed beats an env variable, which beats the default.
4. The tag sets the name in the JSON, so Go names and JSON names can differ. Unexported fields are skipped entirely.
5. Rotate the key (make the old one stop working), then look for misuse, then clean up history. Deleting the file in a new commit doesn't help: the key is still in the history, and bots scan new pushes within minutes.

</details>

## Build it without steps: Chama Contributions

A chama of twelve members in Machakos contributes every month. The treasurer exports an M-Pesa statement as CSV and wants a CLI that tells her who has paid, who's behind, and by how much.

Use only Phases 1 and 2: the standard library, your own packages, no database.

Done when:

- `chama members add NAME PHONE`, `chama members list` and `chama report -month 2026-10 statement.csv` work as subcommands. A bad subcommand or flag prints usage and exits with status 2; a failed run exits with 1.
- Members are saved to a JSON file between runs, at a path from `CHAMA_DATA` (default `chama.json`). The monthly amount comes from `CHAMA_MONTHLY_KSH` or `-monthly`, with the flag winning.
- Config is checked at startup and every problem is reported together, naming the variable or flag.
- Statement rows are matched to members by phone number, normalised (`0712...`, `+254712...` and `254712...` are the same). Rows that match nobody are listed, not silently dropped.
- Parsing errors are your own types: `errors.As` gives the line number, and `errors.Is` matches a sentinel such as `ErrBadAmount`.
- The members store sits behind an interface, and the tests use an in-memory fake.
- Table-driven tests cover the matching, the money (whole cents, never `float64`) and the config rules.
- The work is on a branch in small commits, one idea each, and you've reviewed your own PR before merging it.
