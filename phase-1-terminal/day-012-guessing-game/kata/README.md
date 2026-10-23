# Kata: Binary Search Guesser

A player knows the secret is somewhere from 1 to `N`. Each turn they guess the middle of the range that is still possible, and they're told "too low" or "too high":

```
lo, hi = 1, N
repeat:
    guess = (lo + hi) / 2     (rounded down)
    if guess == secret: stop
    if guess < secret:  lo = guess + 1
    else:               hi = guess - 1
```

For each game, print how many guesses this takes, counting the correct one.

**Input**

- Line 1: `t`, the number of games (1 ≤ t ≤ 10 000)
- The next `t` lines: two numbers `N` and `S`, the top of the range and the secret (1 ≤ S ≤ N ≤ 10⁹)

**Output**

`t` lines, one count per game.

**Example**

```
Input           Output
3               7
100 42          1
1 1             7
100 100
```

For `100 42` the guesses are 50, 25, 37, 43, 40, 41 and 42: seven guesses. 50 is too high, so `hi` becomes 49. 25 is too low, so `lo` becomes 26. And so on.

**Limits:** 2 seconds per test case. Each game takes at most about 30 guesses, so even 10 000 games are fast if you buffer your output.

**Go tip**

Many lines of output are slow if each `fmt.Println` goes straight to the terminal. Wrap stdout in a buffer and flush it once at the end:

```go
out := bufio.NewWriter(os.Stdout)
defer out.Flush()
fmt.Fprintln(out, count)
```

Judge it from the repo root with `go run ./tools/kata 12`.
