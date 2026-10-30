# Day 13: Package It Up

## The brief

Mama Mboga at Gikomba wants a little app that prices a customer's basket. You've written shilling formatting three times already this week, and you're tired of copying it. Today you put it in its own package, `ksh`, and use it from the basket code like any other library.

```
Total(map[string]int{"unga": 2, "sukari": 1}, prices)    →  645, nil
Total(map[string]int{"nyama": 1}, prices)                →  0, no price for "nyama"
Summary(map[string]int{"unga": 2, "sukari": 1}, prices)  →  "3 items, Ksh645", nil
Summary(map[string]int{"sukuma": 1}, prices)             →  "1 item, Ksh20", nil
ksh.Format(3200000)                                      →  "Ksh3,200,000"
```

Here `prices` maps each item to its price in shillings: unga is 230, sukari is 185, sukuma is 20.

Today's folder has two packages:

```
day-013-package-it-up/
  starter/
    basket.go        package packageitup: Total and Summary. The tests check these.
    basket_test.go
    ksh/
      ksh.go         package ksh: Format, plus a helper you write
  solution/
    basket.go        imports .../solution/ksh
    ksh/ksh.go
```

One folder is one package. `starter` is package `packageitup` and `starter/ksh` is package `ksh`. The tests live next to `basket.go` and test `Summary`, which calls `ksh.Format`, so they check both packages at once.

## What you'll use

- **Import paths.** A package is imported by the module path in `go.mod` plus its folder: `github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-013-package-it-up/starter/ksh`. You then call it by its package name: `ksh.Format(...)`.
- **Exported and unexported names.** A name starting with a capital letter, like `Format`, can be used from other packages. A name starting with a small letter, like `group`, is private to its own package. There's no `public` or `private` keyword: the first letter is the rule.
- **Doc comments.** A comment right above `package` or an exported name, starting with that name, is its documentation. `go doc` prints it.
- **`go vet`** reads your code and reports mistakes that compile but are almost surely bugs, like a `Printf` verb that doesn't match its value.

```go
package ksh

// Format writes amount with "Ksh" in front.  ← shows up in go doc
func Format(amount int) string { return "Ksh" + group(amount) }

func group(n int) string { ... } // only code in package ksh can call this
```

```go
fmt.Printf("%d items\n", "three") // go vet: fmt.Printf format %d has arg "three" of wrong type string
```

## Steps

1. Read `starter/ksh/ksh.go`. In the same file, write an unexported helper `group(n int) string` that puts commas into a non-negative number. You wrote this logic on [Day 11](../day-011-receipt-printer/). Import `"strconv"`.
2. Finish `Format`: for a negative amount, return `"-Ksh"` plus `group(-amount)`. Otherwise return `"Ksh"` plus `group(amount)`.
3. Open `starter/basket.go`. Write `Total`: loop over the basket. Look each item up in `prices` with the comma-ok idiom from [Day 7](../day-007-sheng-dictionary/) and return an error if it's missing. Return an error if a quantity is below 1. Add up quantity times price.
4. Write `Summary`: call `Total` and return its error if there is one. Add up the quantities. Use `"item"` when the count is exactly 1 and `"items"` otherwise.
5. Import the `ksh` package in `basket.go` using its full path:

```go
import "github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-013-package-it-up/starter/ksh"
```

   Then build the line with `fmt.Sprintf("%d %s, %s", count, word, ksh.Format(total))`.

6. Try calling `ksh.group(5)` from `basket.go`. Read the error, then delete the line. That error is the point of unexported names.
7. Read your documentation the way other people will see it:

```bash
go doc ./phase-1-terminal/day-013-package-it-up/starter/ksh
go doc ./phase-1-terminal/day-013-package-it-up/starter Summary
```

8. Check both packages with `go vet`. It prints nothing when it finds nothing:

```bash
go vet ./phase-1-terminal/day-013-package-it-up/starter/...
```

9. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-013-package-it-up/starter
```

## The kata: Word Count

Read a block of words and print each distinct word with how many times it appears, most common first. Counting with a map and then sorting the keys is a pattern you'll use in many HackerRank problems. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 13
```

## When you're stuck

- **`undefined: ksh`**: you called `ksh.Format` without importing the package. Add the import from step 5.
- **`"github.com/.../starter/ksh" imported and not used`**: you imported it but haven't called `ksh.Format` yet. Call it, or remove the import until you do.
- **`undefined: ksh.group`**: `group` starts with a small letter, so only package `ksh` can call it. Use `ksh.Format` instead.
- **`no required module provides package github.com/.../ksh`**: the import path is wrong. It must start with the module path from `go.mod`, `github.com/sagebeme/100-Days-of-Go`, followed by the folders.
- **`go doc` prints nothing under a function**: the doc comment must sit directly above the function, with no blank line between them.
- **The total is right but the order of items in your error changes between runs**: maps have no order. That's fine here. The tests only check that there is an error.
- **Still stuck?** Open `solution/basket.go` and `solution/ksh/ksh.go`, read them once, close them, then write your own version from memory.
