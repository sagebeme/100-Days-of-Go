# Kata: Search Suggestions

The job board's search box should suggest job titles as the user types. Given a list of titles and a list of prefixes, print the first three titles in alphabetical order that start with each prefix.

**Input**

- Line 1: `N`, the number of titles (1 ≤ N ≤ 100 000)
- The next `N` lines: one title each, 1 to 20 lowercase letters. All titles are different.
- Then a line with `Q`, the number of prefixes (1 ≤ Q ≤ 100 000)
- The next `Q` lines: one prefix each, 1 to 20 lowercase letters

**Output**

For each prefix, one line: up to three titles that start with it, in alphabetical order and separated by single spaces, or `-` if none do. A title counts as starting with itself.

**Example**

```
Input        Output
6            dentist designer developer
developer    developer devops
designer     data dentist designer
devops       -
data         -
driver
dentist
5
de
dev
d
x
designers
```

`designers` is longer than `designer`, so `designer` doesn't start with it.

**Limits:** 2 seconds per test case.

**Go tip**

Checking every title for every prefix is up to 10^10 steps. Sort the titles once instead. All the titles that start with `p` then sit next to each other, and the first of them is where `p` itself would go:

```go
sort.Strings(titles)
i := sort.SearchStrings(titles, p) // the first title >= p
for ; i < len(titles) && strings.HasPrefix(titles[i], p); i++ { ... }
```

That is O(N log N) to sort and O(log N) per prefix. A trie works too, and is worth trying for practice.

Judge it from the repo root with `go run ./tools/kata 94`.
