# Kata: Almost a Palindrome

A sign writer in Kariakor paints shop names that read the same both ways. Sometimes a brush slips and adds one extra letter. For each word, decide if it is a palindrome after deleting at most one letter.

**Input**

- Line 1: `q`, the number of words (1 ≤ q ≤ 100)
- The next `q` lines: one word each, lowercase letters only, at least 1 letter. The total length of all words is at most 2 × 10^5.

**Output**

For each word, `YES` if deleting zero or one letter makes it a palindrome, otherwise `NO`.

**Example**

```
Input           Output
3               YES
abca            YES
radar           NO
abc
```

`abca` becomes `aca` (or `aba`). `radar` is already a palindrome. `abc` needs two deletions.

**Limits:** 2 seconds per test case.

**Read the constraints.** A word can be 2 × 10^5 letters. Trying every deletion and checking the rest is O(n) × O(n) = O(n²): about 10^10 steps on the longest word. Walk two pointers in from the ends. At the first mismatch, one of those two letters has to go, so check the two possible leftovers. That's O(n).

**Watch out:** `abecbea` is `NO`, but you only see that if you try both deletions at the mismatch. Deleting the left letter and giving up is a common wrong answer.

**Go tip**

Write a helper that checks a range of the string, so you don't copy it:

```go
func isPal(s string, i, j int) bool // is s[i..j] a palindrome?
```

Print with a `bufio.Writer` and `defer out.Flush()`.

Judge it from the repo root with `go run ./tools/kata 29`.
