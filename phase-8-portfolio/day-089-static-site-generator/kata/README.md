# Kata: Tag Balance

Before a generator lets hand-written HTML into a page, it can at least check that the tags are balanced. This is the bracket checker from Day 33, with tag names instead of brackets.

A tag is `<name>` (opening), `</name>` (closing) or `<name/>` (self-closing, which needs no partner), where `name` is 1 or more lower-case letters. Everything else on the line is text. A `<` only ever starts a tag.

**Input**

- Line 1: `q` (1 ≤ q ≤ 1 000)
- The next `q` lines: one snippet each, 0 to 1 000 characters. A line may be empty.

**Output**

One line per snippet:

- `OK` if every opening tag is closed in the right order
- `MISMATCH k` for the first closing tag that doesn't match the innermost open tag (or closes when nothing is open). `k` is the position of its `<`, counting the line's first character as 1.
- `UNCLOSED name` if the line ends with tags still open. `name` is the innermost one.

**Example**

```
Input                          Output
5                              OK
<p>Habari <b>yako</b></p>      MISMATCH 9
<p><b>hi</p></b>               MISMATCH 19
<ul><li>one<li>two</ul>        OK
text only                      OK
<p>line<br/>break</p>
```

In the second line, `</p>` at position 9 arrives while `<b>` is still open. In the third, the two `<li>` tags are never closed, so `</ul>` at position 19 doesn't match.

**Limits:** 2 seconds per test case.

**Go tip**

Lines can be empty, so read them with `bufio.Scanner`, not `fmt.Fscan`, which skips blank lines. Find each tag with `strings.IndexByte(line[i:], '>')` and keep the open names on a slice used as a stack:

```go
stack = append(stack, name)          // push
top := stack[len(stack)-1]           // peek
stack = stack[:len(stack)-1]         // pop
```

Judge it from the repo root with `go run ./tools/kata 89`.
