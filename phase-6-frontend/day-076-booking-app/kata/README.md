# Kata: Seats Together

A family of four wants to sit together on the 58 to Buruburu, and so does a chama heading to a meeting. The seat map has `R` rows of `C` seats: `.` is free and `X` is taken. Group bookings arrive one at a time. Each group of `k` wants `k` free seats next to each other in one row.

For each group, take the first row (from the front) that has `k` free seats in a row, and in that row the leftmost such run. Mark those seats taken and print the row and the first seat's column, both counted from 1. If no row has room, print `SOLD OUT` and change nothing; a smaller group after it may still fit.

**Input**

- Line 1: `R C n` (1 ≤ R ≤ 100, 1 ≤ C ≤ 10, 1 ≤ n ≤ 5 000)
- Next `R` lines: the rows, `C` characters each, from `.` and `X`
- Last line: `n` group sizes (1 ≤ k ≤ 10)

**Output**

`n` lines, each `row column` or `SOLD OUT`, then `free F`: how many seats are still free at the end.

**Example**

```
Input       Output
4 3 5       1 2
X..         2 1
...         3 1
..X         4 2
X..         SOLD OUT
2 3 1 2 2   free 1
```

The first pair takes row 1, seats 2 and 3. The three take all of row 2. The single passenger takes row 3, seat 1, which leaves row 3 with no two free seats side by side, so the next pair goes to row 4. The last pair finds no room.

**Limits:** 2 seconds per test case.

**Go tip:** keep the rows as `[][]byte` so you can change seats in place: `seats[r][c] = 'X'`. Walk each row with a counter of free seats in a row, resetting it at every `X`.

Judge it from the repo root with `go run ./tools/kata 76`.
