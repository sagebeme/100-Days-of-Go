# Day 86: Kiswahili Wordle

## The brief

Mwalimu Atieno teaches Kiswahili at a day school in Kisumu. Her Form 2 class plays Wordle on their phones at lunch, in English. She wants a Kiswahili version they can run in the computer lab: five-letter words, six tries, green and yellow tiles, the same word for the whole class each day, and the English meaning at the end so they learn something even when they lose.

```
Score("mamba", "simba")   →  ..GGG
Score("nyama", "mamba")   →  ..YYG    (two a's guessed, the answer has two: both count)
Score("mtoto", "mdomo")   →  G.G.G    (both o's are green, so nothing is left for the t's)
Daily(24 Oct 2026)        →  "tunda"
Guess("lion")             →  error: neno lazima liwe na herufi tano
Guess("abcde")            →  error: hilo neno halipo kwenye orodha
```

```
Karibu! Kisia neno la herufi 5. Una nafasi 6.
1/6> simba
 S  I  M  B  A   ....G
2/6> nauli
 N  A  U  L  I   YYY..
3/6> tunda
 T  U  N  D  A   GGGGG
Hongera! Umelipata kwa jaribio la 3.
tunda: fruit
```

Acceptance criteria:

- Repeated letters are scored like the real game: greens first, then a yellow only while the answer has an unused copy of that letter.
- Only words in the list are accepted. A bad guess uses up no try.
- The daily word is the same for everyone on the same date.
- The game is playable from `cmd/wordle`, with colours, and readable without them (`NO_COLOR=1`).

The 103 words in `words.go` are given. They are everyday words written with the letters a to z, chosen to avoid `ng'` and other spellings with an apostrophe.

## What you'll use

- Strings as bytes and `strings.ToLower`: [Day 8](../../phase-1-terminal/day-008-hashtag-maker/).
- Maps and the comma-ok idiom for the word list: [Day 7](../../phase-1-terminal/day-007-sheng-dictionary/).
- A game loop on `bufio.Scanner`: [Day 12](../../phase-1-terminal/day-012-guessing-game/), and a state-in, state-out design from [Day 14](../../phase-1-terminal/day-014-lost-in-nairobi/).
- Structs and methods for `Game`: [Day 15](../../phase-2-modelling/day-015-boda-fleet/).
- Sentinel errors and `errors.Is`: [Day 17](../../phase-2-modelling/day-017-failed-transactions/).
- Counting with a fixed array instead of a map, `[26]int`, from [Day 28](../../phase-3-algorithms/day-028-duplicate-receipts/).

```go
var left [26]int
left['k'-'a']++ // one unused k in the answer
```

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `Words` returns every word in `Meanings` in alphabetical order. `IsWord` is exact: `SIMBA` is not a word until it's normalized.
2. `Normalize` trims spaces and lower-cases.
3. `Score` follows its doc comment, and `Pattern` turns marks into `G`, `Y` and `.`.
4. `Daily` counts whole days from 1 January 2026 in `t`'s own time zone and indexes `Words()`, wrapping round. Dates before 2026 work too.
5. `NewGame` and `Game.Guess`: errors for the wrong length, unknown words and guessing after the end, none of which uses a try.
6. `Won`, `Over`, `Answer`, a `Turns` that returns a copy, and `Letters` for the keyboard.
7. `cmd/wordle` plays today's word, or a random one with `-random`. It prints coloured tiles, the keyboard, and the meaning at the end. Without colour, it prints the pattern after the tiles.
8. Run the tests and play it:

```bash
go test ./phase-8-portfolio/day-086-kiswahili-wordle/starter
go run ./phase-8-portfolio/day-086-kiswahili-wordle/starter/cmd/wordle
```

Stretch goals:

- Hard mode: every guess must use the greens and yellows already found.
- Share text like the real game, with `G`, `Y` and `.` rows and no letters, so it doesn't spoil the word.
- Stats saved to a JSON file ([Day 23](../../phase-2-modelling/day-023-save-load/)): games played, win streak, guesses per win.
- Separate lists for answers and allowed guesses, so rarer words can be guessed but are never the answer.

## The kata: Still Possible

Given the guesses and patterns so far, count the words that could still be the answer. It's today's `Score` used as a filter, and a favourite follow-up question. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 86
```

## When you're stuck

- **`mamba` against `nyama` gives `YYY.G`, not `YY..G`**: you marked the second `m` yellow too. `nyama` has only one `m`, and the first `m` already used it. Count the answer's spare letters once, before the yellow pass, and take one away for each yellow.
- **`mtoto` against `mdomo` gives `GYGYG`**: you checked `strings.Contains(answer, ...)`. That ignores how many copies are left. Use the `[26]int` count.
- **A green gets turned yellow**: do all the greens in a first pass, then the yellows in a second. In one pass, an early yellow can use up the copy a later green needed.
- **`Daily` changes word at 3 a.m. in Nairobi**: you counted days in UTC. Take the year, month and day from `t` as it is, then build a UTC midnight from them before you subtract.
- **`Daily` panics for 2025**: `-1 % 103` is `-1` in Go. Add the length when the remainder is negative.
- **`index out of range` in `Score`**: a guess got through that isn't 5 letters. Check the length in `Guess` before scoring.
- **The colours show as `[42m`**: your terminal doesn't do ANSI codes. Run with `NO_COLOR=1`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Give it its own repo and a README with a screenshot of a coloured game in the terminal, or a short GIF recorded with `vhs` or `asciinema`.
- Release binaries for Windows, macOS and Linux with GoReleaser, so a school lab can run it without Go installed. Link the latest release.
- Explain in the README how the word list was chosen and how to add words. Invite a Kiswahili teacher to review it.
- Write "What I'd do next": hard mode, saved stats, a web version.
- Add GitHub Actions running `go vet` and `go test`, with the CI badge at the top.
