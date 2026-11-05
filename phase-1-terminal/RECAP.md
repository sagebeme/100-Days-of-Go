# Phase 1 recap: Terminal Go

Days 1 to 14. You went from printing one M-Pesa SMS to a text adventure with rooms, exits and a command loop, and every step had tests.

## What the phase covered

| Days | Topic |
|---|---|
| 1 to 2 | values, constants, `fmt.Sprintf`; integers, floats and converting between them |
| 3 to 5 | `if`, `else` and `switch`; `for` loops and accumulators; functions, several return values, errors as values |
| 6 to 8 | slices, `append` and `range`; maps and the comma-ok idiom; strings, bytes, runes and the `strings` package |
| 9 to 11 | checking input byte by byte; sorting with `sort.Slice`; columns with `fmt` and `text/tabwriter` |
| 12 to 13 | reading stdin with `bufio`, `math/rand`; packages, exported names, `go doc`, `go vet` |
| 14 | the capstone: Lost in Nairobi CBD, a text adventure driven by a map of rooms |

## Key ideas

1. **Money is whole numbers.** Keep shillings or cents in an `int`; floats can't hold most decimal fractions exactly.
2. **Errors are values.** A function that can fail returns `(result, error)`, and the caller checks `err != nil` straight away. No exceptions, no surprises.
3. **A slice is a window onto an array.** `append` may write into the array you share with another slice, or copy to a new one. Copy when you need your own.
4. **Missing map keys give the zero value.** Use `v, ok := m[k]` when "missing" and "zero" mean different things.
5. **A string is bytes, not characters.** `len` counts bytes; `range` and `[]rune` work in runes, so `ũ` counts once.
6. **Map order is random.** Sort the keys whenever output must be the same every run, and tests need that.
7. **Keep logic out of `main`.** A function that takes input and returns output is easy to test; `main` only reads, calls and prints.
8. **Tests are the spec.** `go test` tells you when you're done; `go vet` catches mistakes the compiler allows.

## Self-test

1. Predict the output:

   ```go
   s := "Habari, Wanjikũ"
   fmt.Println(len(s), len([]rune(s)))
   a := []int{1, 2, 3}
   b := a[:2]
   b = append(b, 99)
   fmt.Println(a, b)
   ```

2. What do `7/2`, `7%2` and `-7/2` give in Go?
3. `m := map[string]int{"ugali": 50}`. What do `v, ok := m["chapati"]` give, and why would `m["chapati"] == 0` be the wrong test for "not on the menu"?
4. Why does Day 14's `Describe` sort the exits before printing them?
5. A function reads a number from the user and can fail. What should it return, and what should the caller do first?

<details>
<summary>Answers</summary>

1. `16 15`: the `ũ` takes two bytes. Then `[1 2 99] [1 2 99]`: `b` had room in `a`'s array, so `append` overwrote `a[2]`.
2. `3`, `1` and `-3`: integer division throws the fraction away, rounding towards zero.
3. `0` and `false`. A dish could really cost 0 (a free sample), so only `ok` tells you whether the key is there.
4. Map iteration order changes between runs, so the same room would describe its exits differently, and the tests couldn't check the text.
5. The number and an `error`. The caller checks `err != nil` before using the number.

</details>

## Build it without steps: Boda Trip Log

Juma rides a boda boda in Kisumu. Every night he types his trips into a phone notes app, one per line, like `Kondele,Milimani,4.5,120`: from, to, kilometres, fare in shillings. He wants a program that reads those lines from stdin and prints a tidy summary.

Use only what Phase 1 taught: no structs, no third-party packages. Decide the functions yourself, put them in their own package, and write the tests first.

Done when:

- `go run ./cmd/bodalog < trips.txt` prints a table, lined up with `text/tabwriter`: each trip, then the total kilometres (one decimal place), the total fare and the average fare per kilometre.
- It also prints each destination with how many trips ended there, busiest first, ties in alphabetical order, so the output is the same on every run.
- Blank lines are skipped. A bad line (wrong number of fields, a fare that isn't a whole number, negative kilometres) is reported with its line number on stderr and skipped; the rest still count.
- Place names are trimmed and compared without caring about case: `kondele` and `Kondele ` are the same stage.
- Fares are added as integers, never floats.
- `go test` covers every rule above with tests that don't touch stdin, and `go vet` and `gofmt -l` print nothing.
