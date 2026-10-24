# Day 20: Go Modules

## The brief

Your chama app uses a library someone else wrote. They release `v1.4.1` with a fix, then `v1.5.0-rc.1` to try out a feature, then `v2.0.0`, which renames half the functions. Which one should `go get -u` move you to? (`v1.4.1`. The rc is a pre-release, and `v2` is a different module.)

Today you learn how Go modules manage dependencies, and you write the version logic underneath: parse a semantic version, compare two of them the way the spec says, and pick the upgrade `go get -u` would pick.

```
Parse("v2.0.0-rc.1")                  →  Version{Major: 2, Minor: 0, Patch: 0, Pre: "rc.1"}, nil
Parse("v1.02.0")                      →  error: "02" has a leading zero
Compare(v1.10.0, v1.9.0)              →  1      (numbers, not strings)
Compare(v1.0.0-rc.1, v1.0.0)          →  -1     (a pre-release comes before its release)

Upgrade("v1.2.0", []string{"v1.3.0", "v1.4.1", "v1.5.0-rc.1", "v2.0.0"})
                                      →  "v1.4.1"
ModulePath("github.com/sagebeme/pesa", 2)
                                      →  "github.com/sagebeme/pesa/v2"
```

## What you'll use

### Semantic versions

A version is `vMAJOR.MINOR.PATCH`, with a promise attached:

| You bump | When | Users can upgrade |
|---|---|---|
| PATCH, `v1.4.0` → `v1.4.1` | a bug fix, nothing else changes | without thinking |
| MINOR, `v1.4.1` → `v1.5.0` | new features, old code still compiles | safely |
| MAJOR, `v1.5.0` → `v2.0.0` | something breaks: a renamed function, a removed field | only by changing their code |

`v0.x.y` means "not stable yet": anything may change. A **pre-release** like `v1.5.0-rc.1` comes *before* `v1.5.0`. Go never picks one unless you ask for it by name.

**Go's rule for majors:** from `v2` on, the major version is part of the import path. `github.com/jackc/pgx/v5` and `github.com/jackc/pgx/v4` are two different modules, and a program can even use both. That's why a `go get -u` never jumps you to a new major.

### go.mod and go.sum

Open `go.mod` at the root of this repo:

```
module github.com/sagebeme/100-Days-of-Go

go 1.26.0

require (
	github.com/jackc/pgx/v5 v5.11.0
	golang.org/x/crypto v0.58.0
	...
)

require (
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	...
)
```

- `module` is this module's path. Every import of a package in the repo starts with it, which is why Day 13 imported `github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-013-package-it-up/solution/ksh`.
- `go` is the oldest Go version that can build it.
- `require` lists each dependency and the **minimum** version needed. `// indirect` marks one your code doesn't import directly; a dependency of a dependency needs it.
- `go.sum` holds a checksum for each module version. The go command checks every download against it, and against Google's public checksum database, so nobody can quietly swap the code behind a version. Commit both files.

Some versions look like `v0.0.0-20240606120523-5a60cdf6a761`. That's a **pseudo-version**: a commit that has no tag, named by its time and commit hash. It's still a valid semver pre-release, so your `Parse` reads it.

### The commands

Try these in a scratch folder **outside this repo**, so you don't change the course's `go.mod`:

```bash
mkdir ~/scratch-pesa && cd ~/scratch-pesa
go mod init example.com/pesa          # creates go.mod
go get github.com/google/uuid@v1.6.0  # add a dependency at an exact version
go get github.com/google/uuid@latest  # move to the newest release
go get -u ./...                       # upgrade everything to the newest minor or patch
go get -u=patch ./...                 # patch releases only
go mod tidy                           # add what your code imports, remove what it doesn't
go list -m all                        # every module in the build, with its version
go list -m -u all                     # the same, plus available upgrades in [brackets]
go mod why -m github.com/google/uuid  # which of your packages needs it
```

Run `go mod tidy` before every commit that touches imports. CI often fails a build when `go mod tidy` would change something.

**How Go picks versions**: if your code needs `x v1.2.0` and another dependency needs `x v1.3.0`, Go uses `v1.3.0`: the *lowest* version that satisfies everyone. This is called minimal version selection. It never picks a newer version just because one exists, so a build today and a build next year use the same code.

**`replace`** points a module at somewhere else, most often a local copy you're fixing:

```
replace github.com/someone/mpesa => ../mpesa
```

It only applies when your module is the one being built, never when someone else depends on yours. Remove it before you release.

**Vendoring**: `go mod vendor` copies every dependency into a `vendor/` folder, and `go build` then uses that copy. Some companies do this so builds don't need the network, or so reviewers see dependency changes in the diff. Most open-source projects don't, because the module cache and `go.sum` already make builds repeatable.

## Why it matters at work

Every dependency is code you ship but didn't write. Code review should ask about a new one: is it maintained, how big is it, what does it pull in (`go mod graph`), and is the standard library enough? A diff to `go.mod` and `go.sum` deserves a look, not a rubber stamp. Keep upgrades small and regular. A patch bump with green tests is a five-minute PR. Two years of skipped updates is a week of work, often done in a hurry after a security alert. `govulncheck ./...` (`go install golang.org/x/vuln/cmd/govulncheck@latest`) tells you whether your code calls a function with a known vulnerability.

## Steps

1. Open `starter/semver.go`. Read the doc comments: they are the rules.
2. Write `Parse`. Use `strings.CutPrefix(s, "v")` for the `v`. Reject a `+` anywhere. `strings.Cut(rest, "-")` splits off the pre-release, at the *first* `-` only, because pre-releases may contain `-` too. Then `strings.Split` the core on `.` and check you have exactly three parts.
3. Write a small helper that reads one number: not empty, all digits, no leading zero unless it's just `"0"`. Then check each pre-release identifier the same way the comment describes.
4. Write `String`. It's the reverse of `Parse`.
5. Write `Compare`. Major, minor, patch first. Then the pre-release rules: a release beats any pre-release of the same version; otherwise compare identifier by identifier. `TestCompareOrder` uses the list from the semver spec itself.
6. Write `Upgrade` with `Parse` and `Compare`: keep the best version so far, starting at `current`.
7. Write `ModulePath`.
8. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-020-go-modules/starter
```

9. In your scratch module, run `go list -m all` and `go list -m -u all`. Find a dependency with an upgrade and check that your `Upgrade` would pick the same one.

## The kata: Compare Versions

Compare loose version numbers like `1.01` and `1.001`, where missing parts count as zero. It's LeetCode's "Compare Version Numbers", a common interview warm-up. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 20
```

## When you're stuck

- **`TestCompare/v1.10.0_vs_v1.9.0` fails**: you compared strings. `"10" < "9"` as text. Compare the numbers.
- **`v1.0.0-x-y-z.0` doesn't parse**: you split on every `-`. `strings.Cut` splits only at the first one.
- **`Parse("v1.2.3-")` doesn't error**: `strings.Cut` reports that it found the `-`, and the pre-release after it is `""`. An empty pre-release is an error.
- **`v1.0.0-beta.11` sorts before `v1.0.0-beta.2`**: digit-only identifiers compare as numbers. Since leading zeros are banned, the longer one is bigger, and same-length ones compare as text.
- **`Upgrade` returns `v2.1.0` for `v1.2.0`**: skip versions with a different major.
- **`go: go.mod file not found in current directory or any parent directory`** in your scratch folder: run `go mod init example.com/pesa` first.
- **`go: updates to go.mod needed`**: run `go mod tidy`, or the exact `go get` it suggests.
- **Still stuck?** Open `solution/semver.go`, read it once, close it, then write your own version from memory.
