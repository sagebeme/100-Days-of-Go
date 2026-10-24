# Kata: Still Possible

A Wordle helper is the classic follow-up interview question: given what the game has told you so far, which words could still be the answer? A word is still possible when scoring every past guess against it gives exactly the pattern the game showed.

Scoring is today's rule. `G` is the right letter in the right place. `Y` is a letter that's in the answer somewhere else, and `.` is a letter that isn't. Greens are marked first. A letter is then `Y` only while the answer has an unused copy of it, so `mtoto` against `mdomo` is `G.G.G`: both `o`s are green and neither `t` is anything.

**Input**

- Line 1: `n` (1 ≤ n ≤ 20 000)
- The next `n` lines: a word of 5 letters from `a` to `z`. No word appears twice.
- The next line: `m` (0 ≤ m ≤ 6)
- The next `m` lines: a guess (5 letters, not necessarily in the list) and its 5-character pattern of `G`, `Y` and `.`

**Output**

- Line 1: how many of the `n` words are still possible
- Then the first 10 of them in alphabetical order, one per line (all of them if there are fewer than 10)

**Example**

```
Input       Output
6           1
simba       simba
mamba
tembo
punda
nyama
panda
1
mamba ..GGG
```

`mamba` against `panda` gives `.G..G`, not `..GGG`, so `panda` is out. Only `simba` fits.

**Limits:** 2 seconds per test case. 20 000 words times 6 guesses is 120 000 scorings of 5 letters: easy.

**Go tip**

Score with a `[26]int` of unused answer letters, not a map. It lives on the stack and is reset for free each call:

```go
var left [26]int
left[answer[i]-'a']++
```

Judge it from the repo root with `go run ./tools/kata 86`.
