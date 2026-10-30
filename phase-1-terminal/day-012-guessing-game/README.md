# Day 12: Number Guessing Game

## The brief

Your cousin's matatu has a screen in the back, and the owner wants a game for passengers stuck in Thika Road traffic. The matatu picks a secret number from 1 to 100, the passenger types guesses, and the screen says "Too low." or "Too high." until they get it or run out of tries. Write the game so it can be played in a terminal, and so a test can play it too.

```
Hint(42, 50)                  →  "Too high."
NewSecret(r, 100)             →  a number from 1 to 100
Play(strings.NewReader("50\n25\n42\n"), &out, 42, 100, 7)  →  true, 3
```

and `out` now holds:

```
Guess a number from 1 to 100. You have 7 tries.
> Too high.
> Too low.
> Correct! Guesses used: 3.
```

A game that reads the keyboard and picks a random number is hard to test: you can't type into a test, and you can't predict the secret. The fix is to pass both in. `Play` reads from any `io.Reader` and writes to any `io.Writer`, and it gets the secret as an argument. In the terminal you pass `os.Stdin` and `os.Stdout`. In a test you pass a string and a `strings.Builder`.

## What you'll use

- `io.Reader` means "anything you can read from" and `io.Writer` means "anything you can write to". `os.Stdin`, a file and `strings.NewReader("...")` are all readers. `os.Stdout` and `*strings.Builder` are writers. [Day 16](../../phase-2-modelling/day-016-payment-methods/) explains how this works.
- `bufio.NewScanner(in)` reads one line at a time. `sc.Scan()` returns `false` when the input runs out, and `sc.Text()` is the line.
- `fmt.Fprintf(out, ...)` and `fmt.Fprintln(out, ...)` are `Printf` and `Println` that write to `out`.
- `strconv.Atoi(s)` turns `"42"` into `42`, and returns an error for `"abc"`.
- `math/rand/v2`. `rand.New(rand.NewPCG(seed1, seed2))` makes a generator. The same seeds always give the same numbers, which is how a test can rely on it. `r.IntN(n)` returns a number from 0 to n−1.

```go
sc := bufio.NewScanner(strings.NewReader("12\nabc\n"))
for sc.Scan() {
	n, err := strconv.Atoi(sc.Text())
	fmt.Println(n, err) // 12 <nil>, then 0 strconv.Atoi: parsing "abc": invalid syntax
}

r := rand.New(rand.NewPCG(1, 2))
fmt.Println(r.IntN(6) + 1) // a dice roll, and the same one every run
```

## Steps

1. Open `starter/game.go`. Add `"bufio"`, `"fmt"`, `"strconv"` and `"strings"` to the imports as you use them.
2. `Hint`: compare the guess with the secret and return `"Too low."`, `"Too high."` or `"Correct!"`.
3. `NewSecret`: return 0 if `max` is below 1. Otherwise return `r.IntN(max) + 1`.
4. `Play`: make a scanner on `in`. Print the welcome line with `fmt.Fprintf(out, ...)`. Copy the exact text from the brief.
5. Loop while `used < tries`. Print `"> "` with no newline. If `sc.Scan()` returns false, print `"\nBye! It was 5.\n"` (with the real secret) and return `false, used`.
6. Trim the line and convert it with `strconv.Atoi`. On an error, print `Numbers only, please.` and `continue`. If the number is below 1 or above `max`, print `From 1 to 100, please.` (with the real max) and `continue`. Neither uses up a try.
7. Add one to `used`. On the right answer, print `Correct! Guesses used: 3.` and return `true, used`. Otherwise print the hint.
8. After the loop, print `Out of tries. It was 7.` and return `false, used`.
9. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-012-guessing-game/starter
```

10. Play it for real: `go run ./phase-1-terminal/day-012-guessing-game/starter/cmd/play`. Press Ctrl+D to give up.

## The kata: Binary Search Guesser

The best way to play your game is to guess the middle every time. Given the range and the secret, count how many guesses that takes. This is binary search, the idea behind half of all interview questions on sorted data. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 12
```

## When you're stuck

- **`"io" imported and not used`** or **`"math/rand/v2" imported and not used`**: the starter needs these for the function signatures. Don't delete them.
- **The test shows `"> \n"` missing**: print the prompt before you call `sc.Scan()`, and print the `\n` before `Bye!` yourself, since the player never pressed Enter.
- **`" 42 "` gives `Numbers only, please.`**: `strconv.Atoi` doesn't skip spaces. Use `strings.TrimSpace` first.
- **`used` is too high**: you added one before checking that the line was a number in range. Only a real guess counts.
- **`NewSecret` gives 0 or never gives 10**: `r.IntN(10)` returns 0 to 9. Add one.
- **The play command sits there doing nothing**: it's waiting for you. Type a number and press Enter.
- **Still stuck?** Open `solution/game.go`, read it once, close it, then write your own version from memory.
