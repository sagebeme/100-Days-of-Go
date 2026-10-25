# Kata: Leak Scanner

Someone pasted Twende's logs into a public chat to ask for help, and some lines had secrets in them. Before logs leave the building, they go through a scanner that blanks out anything that looks like a secret. Write it.

A line is a list of words separated by single spaces. Go through the words left to right and replace a word, or part of it, with `[REDACTED]` by these rules:

1. **Assignments.** A word containing `=` is `NAME=VALUE`, split at the first `=`. If `NAME`, in lower case, contains `password`, `secret`, `token` or `key`, and `VALUE` isn't empty, the word becomes `NAME=[REDACTED]`.
2. **Live keys.** A word that starts with `sk_live_` and is longer than 8 characters becomes `[REDACTED]`.
3. **Bearer tokens.** A word that comes straight after a word that was exactly `Bearer` in the input becomes `[REDACTED]`, whatever it is. So `Bearer Bearer x` has two redactions.

Apply rule 3 first: a word after `Bearer` is redacted whole and no other rule looks at it. Then rule 1, then rule 2. Each word is redacted at most once. Words that are redacted count one each.

**Input**

- Line 1: `n`, the number of log lines (1 ≤ n ≤ 10 000)
- Then `n` lines, each 1 to 500 characters of printable ASCII with no leading, trailing or double spaces

**Output**

The `n` lines with secrets replaced, then one more line: `redacted N`, the number of words redacted.

**Example**

```
Input
4
level=INFO msg=login user=wanjiku
db_password=hunter2 host=10.0.3.7
GET /trips Authorization: Bearer eyJhbGciOi.payload.sig
daraja=sk_live_abc123 fallback sk_live_9xY KEY= token=

Output
level=INFO msg=login user=wanjiku
db_password=[REDACTED] host=10.0.3.7
GET /trips Authorization: Bearer [REDACTED]
daraja=sk_live_abc123 fallback [REDACTED] KEY= token=
redacted 3
```

`daraja=sk_live_abc123` is an assignment whose name doesn't match rule 1, and the word doesn't *start* with `sk_live_`, so a real scanner would miss it too. That's why the real fix is never logging secrets in the first place (`slog.LogValuer`, Day 83). `KEY=` and `token=` have empty values. The rules also have false alarms: `MonkeyBusiness=1` contains `key`, so it's redacted. Scanners prefer too much to too little.

**Limits:** 2 seconds per test case.

**Go tip:** read lines with `bufio.Scanner`; after reading `n` with `sc.Scan()` and `strconv.Atoi(sc.Text())`, each `sc.Scan()` gives one log line. `strings.Fields` splits a line into words, `strings.Cut(w, "=")` splits at the first `=`, and `strings.ToLower`, `strings.Contains` and `strings.HasPrefix` do the checks. Join the words back with `strings.Join(words, " ")`.

Judge it from the repo root with `go run ./tools/kata 83`.
