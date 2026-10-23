# Kata: Grid Walk

A boda rider is on a grid of CBD blocks. `.` is an open block, `#` is a building and `S` is where the rider starts (an open block). They follow a string of moves: `N` is up, `S` is down, `E` is right and `W` is left. A move into a building or off the edge of the grid does nothing, and counts as blocked.

Print where the rider ends up and how many moves were blocked.

**Input**

- Line 1: `R` and `C`, the number of rows and columns (1 ≤ R, C ≤ 50)
- The next `R` lines: `C` characters each, from `.`, `#` and exactly one `S`
- The last line: the moves, a string of 1 to 100 000 characters from `N`, `S`, `E` and `W`

**Output**

- Line 1: the final row and column, numbered from 1, with the top-left block as `1 1`
- Line 2: the number of blocked moves

**Example**

```
Input           Output
3 4             2 4
S..#            3
.#..
....
EEENESE
```

From `1 1`, the first two `E` moves reach `1 3`. The third `E` hits the building at `1 4`: blocked. `N` would leave the grid: blocked. `E` hits the building again: blocked. `S` reaches `2 3` and `E` reaches `2 4`. The rider ends at `2 4` after 3 blocked moves.

**Limits:** 2 seconds per test case.

**Go tip**

Read each grid row as a string with `fmt.Fscan`. `grid[r][c]` is then a byte you can compare with `'#'`. Keep the moves in a small map so the loop stays short:

```go
dr := map[byte]int{'N': -1, 'S': 1}
dc := map[byte]int{'E': 1, 'W': -1}
nr, nc := r+dr[m], c+dc[m] // a missing key gives 0
```

Judge it from the repo root with `go run ./tools/kata 14`.
