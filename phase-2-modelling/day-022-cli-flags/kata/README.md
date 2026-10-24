# Kata: Flag Parser

Write the heart of a command-line flag parser. You get the flags a program defines and the words typed after its name. Print each flag's final value and the arguments left over, or the first error.

These are the rules Go's `flag` package follows, slightly simplified:

1. Go through the words in order. Stop parsing flags at the first word that is `--` (skip it), at a word that doesn't start with `-`, or at the word `-` on its own. Everything from there on is a leftover argument.
2. A flag word starts with `-` or `--`. Remove that. If what's left is empty or starts with `-` or `=`, the error is `bad flag syntax: WORD` (the whole word).
3. If what's left contains `=`, the name is before the first `=` and the value is after it. Otherwise it's just a name.
4. A name that isn't defined: `flag provided but not defined: -NAME`.
5. A `bool` flag with no `=` becomes `true`. With `=`, the value must be one of `1 t T TRUE true True 0 f F FALSE false False`, or the error is `invalid boolean value "VALUE" for -NAME`. A bool never takes the next word as its value.
6. An `int` or `string` flag with no `=` takes the **next word** as its value, even if it starts with `-`. If there is no next word: `flag needs an argument: -NAME`.
7. An `int` value is an optional `+` or `-` and then one or more digits, or the error is `invalid value "VALUE" for flag -NAME`. Print ints without a sign for positive numbers and without leading zeros: `+007` is `7`.

A flag given twice keeps its last value.

**Input**

- Line 1: `n`, the number of defined flags (1 ≤ n ≤ 50)
- The next `n` lines: a name, a type (`int`, `string` or `bool`) and a default value. Names are lowercase letters and digits.
- Then `m`, the number of words (0 ≤ m ≤ 20 000)
- The next `m` lines: one word each. Words have no spaces and are 1 to 40 characters.

Int values have at most 15 digits.

**Output**

If there's an error, print one line: `error: ` and the message. Otherwise print `name=value` for every flag in the order defined, then a line `args:` followed by a space and each leftover argument.

**Example**

```
Input
4
total int 0
members int 0
dry bool false
name string chama
5
-total
12005
--members=5
-dry
report.txt

Output
total=12005
members=5
dry=true
name=chama
args: report.txt
```

`-total` takes the next word, `12005`. `--members=5` carries its own value. `-dry` is a bool, so it becomes `true` without eating `report.txt`, which doesn't start with `-` and so ends the flags.

**Limits:** 2 seconds per test case.

**Go tip**

Keep an index `i` into the words instead of a `range` loop, because a flag may eat the next word too. `strings.Cut(rest, "=")` gives you the name, the value and whether there was an `=`. Store each flag's type and current value in two maps, and keep the names in a slice for the output order.

Judge it from the repo root with `go run ./tools/kata 22`.
