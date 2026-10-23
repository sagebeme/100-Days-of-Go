# Kata: Sheng Palindromes

A palindrome reads the same forwards and backwards. For each line, decide whether it is a palindrome once you ignore case and keep only letters and digits. So `Was it a car or a cat I saw?` counts, and so does `Kayak`.

**Input**

- Line 1: `n`, the number of lines to check (1 ≤ n ≤ 2 000)
- The next `n` lines: one line of text each, up to 1 000 characters. Text is plain ASCII and may contain spaces and punctuation. A line can be empty.

**Output**

For each line, print `YES` if it is a palindrome and `NO` if it isn't. A line with no letters or digits at all is a palindrome.

**Example**

```
Input                          Output
3                              YES
Kayak                          NO
Habari                         YES
Was it a car or a cat I saw?
```

**Limits:** 2 seconds per test case.

**Reading whole lines in Go**

`fmt.Fscan` stops at spaces, so it can't read a sentence. Use a `bufio.Scanner`, which hands you one line at a time:

```go
sc := bufio.NewScanner(os.Stdin)
sc.Scan()
n, _ := strconv.Atoi(sc.Text())
for i := 0; i < n; i++ {
	sc.Scan()
	line := sc.Text()
	_ = line // check it here
}
```

Compare from both ends with two indexes, `i` from the left and `j` from the right, moving inwards. That avoids building a reversed copy.

Judge it from the repo root with `go run ./tools/kata 8`.
