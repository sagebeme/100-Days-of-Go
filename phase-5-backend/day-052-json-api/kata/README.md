# Kata: JSON Brackets

Before a JSON parser looks at values, it has to know that every `{` and `[` is closed, in the right order. Write that first check, the way a parser does it: brackets inside a string don't count.

For each line, scan from left to right:

- A `"` starts a string, and the next `"` that isn't escaped ends it. Inside a string, a `\` escapes the character after it, so `\"` doesn't end the string and `\\` is one backslash. Nothing inside a string is a bracket.
- Outside strings, `{` and `[` open, and `}` and `]` close. A closer must match the most recent bracket that is still open.
- Every other character is ignored. This kata checks brackets only, not the whole of JSON.

Print `YES` if the line is balanced. Otherwise print `NO` and the position (counting from 1) of the first problem: a closer with nothing open or the wrong kind open. If the end of the line comes with a string or a bracket still open, the position is the line's length plus 1.

**Input**

- Line 1: `n`, the number of lines (1 ≤ n ≤ 200)
- The next `n` lines: each up to 1 000 characters of printable ASCII, possibly empty

**Output**

`n` lines, each `YES` or `NO k`.

**Example**

```
Input                    Output
5                        YES
{"seats": [1, 2, 3]}     YES
{"note": "seat [3"}      NO 12
{"a": [1, 2}             NO 8
{"a": 1                  NO 1
}
```

Line 2's `[` is inside a string, so it doesn't count. Line 3 closes with `}` at position 12 while `[` is still open. Line 4 has 7 characters and ends with `{` open, so the answer is 8.

**Limits:** 2 seconds per test case.

**Go tip**

A stack is a slice you append to and cut from the end (Day 33). Keep a bool for "inside a string" and another for "the last character was a backslash":

```go
stack := []byte{}
stack = append(stack, '{')        // push
top := stack[len(stack)-1]        // peek
stack = stack[:len(stack)-1]      // pop
```

Read lines with a `bufio.Scanner`, because a line can be empty or full of spaces.

Judge it from the repo root with `go run ./tools/kata 52`.
