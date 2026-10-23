# Day 8: Hashtag Maker

## The brief

A Nairobi radio station runs its promos on X and TikTok. The social media intern types the campaign name in plain words, and somebody has to turn it into a hashtag, put the presenter's initials on the poster, and cut long captions short for the preview. Write three functions that do it.

```
Hashtag("nairobi tech week")       →  "#NairobiTechWeek"
Hashtag("  MASHUJAA day  2026! ")  →  "#MashujaaDay2026"
Hashtag("café ya mama")            →  "#CaféYaMama"
Initials("Ngũgĩ wa Thiong'o")      →  "NWT"
Truncate("Habari za asubuhi", 6)   →  "Habari..."
Truncate("café latte", 4)          →  "café..."
```

The last two lines are the trap. `é` and `ũ` are one letter on screen but two bytes in memory. Cut by bytes and you slice a letter in half.

## What you'll use

- A `string` is a row of bytes. `len(s)` counts bytes, not letters.
- A `rune` is one character. `for _, r := range s` walks a string rune by rune, whatever the bytes.
- `[]rune(s)` turns a string into a slice of runes you can index and slice. `string(runes)` turns it back.
- The `strings` package: `strings.Fields` splits on any run of spaces, `strings.ToUpper` and `strings.ToLower` change case, and `strings.Builder` builds a string piece by piece.
- The `unicode` package: `unicode.IsLetter(r)`, `unicode.IsDigit(r)`, `unicode.ToUpper(r)`. The `unicode/utf8` package: `utf8.RuneCountInString(s)` counts letters.

```go
s := "café"
len(s)                       // 5: é takes two bytes
utf8.RuneCountInString(s)    // 4
strings.Fields("  a  b ")    // ["a" "b"]

var b strings.Builder
for _, r := range "kip" {
	b.WriteRune(unicode.ToUpper(r))
}
b.String()                   // "KIP"
```

## Steps

1. Open `starter/hashtag.go`. Don't change the function names or parameters.
2. Add the imports you need under the `package` line: `"strings"`, `"unicode"` and `"unicode/utf8"`. Add them as you use them, because an unused import is a compile error.
3. `Hashtag`: loop over `strings.Fields(phrase)`. For each word, loop over its runes. Skip a rune that is not a letter and not a digit. Write the first rune you keep in upper case and the rest in lower case into a `strings.Builder`.
4. Still in `Hashtag`: if the builder is empty, return `""`. Otherwise return `"#"` plus the builder's string.
5. `Initials`: for each word from `strings.Fields`, take its first rune and write it in upper case. `utf8.DecodeRuneInString(word)` returns the first rune and its size, or use a `range` loop and `break` after one turn.
6. `Truncate`: if `utf8.RuneCountInString(s) <= max`, return `s` unchanged. Otherwise convert to `[]rune`, keep the first `max`, turn them back into a string and add `"..."`.
7. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-008-hashtag-maker/starter
```

## The kata: Sheng Palindromes

A palindrome reads the same both ways, like "Kayak". Decide for each line whether it is one, ignoring case, spaces and punctuation. Reversing and comparing strings is one of the most common interview warm-ups. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 8
```

## When you're stuck

- **`"strings" imported and not used`**: you added an import before writing the code that uses it. Use it, or delete the line for now.
- **`cannot use r (variable of type rune) as string value`**: a rune is a number, not a string. Write it with `b.WriteRune(r)`, or convert with `string(r)`.
- **The test shows `"caf\xc3..."`**: you sliced bytes with `s[:max]`. Convert to `[]rune(s)` first and slice that.
- **`#Nairobitechweek` or `#NAIROBITECHWEEK`**: you changed case for the whole phrase at once. Only the first kept rune of each word goes to upper case; every other rune goes to lower case.
- **`#SemaNg'ware`**: the apostrophe isn't a letter or digit, so skip it with `unicode.IsLetter` and `unicode.IsDigit`.
- **Still stuck?** Open `solution/hashtag.go`, read it once, close it, then write your own version from memory.
