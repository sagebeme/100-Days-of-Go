# Kata: Expand the .env

Many `.env` loaders let one value use another: `DATABASE_URL=postgres://${DB_USER}@${DB_HOST}/chama`. Read a `.env` file and print every variable with its references filled in.

Rules:

- Process the lines from top to bottom. Skip blank lines and lines whose first non-space character is `#`.
- Every other line has a `=`. The key is the text before the first `=` and the value is everything after it, both with spaces trimmed from the ends.
- In the value, replace each `${NAME}` with the value `NAME` has **at that point** in the file, or with nothing if it isn't set yet. `${}` becomes nothing. A `${` with no `}` after it is plain text, and so is a `$` that isn't followed by `{`.
- A key that appears again replaces its old value. The new value may use the old one: `A=${A}x`.

**Input**

A `.env` file of up to 5 000 lines, each up to 200 characters. Keys are letters, digits and `_`.

**Output**

Every key that was set, sorted by key (byte order), as `KEY=value`, one per line.

**Example**

```
Input
# chama service
DB_USER=chama
DB_HOST=localhost
DATABASE_URL=postgres://${DB_USER}@${DB_HOST}:5432/chama
GREETING=Habari ${NAME}

Output
DATABASE_URL=postgres://chama@localhost:5432/chama
DB_HOST=localhost
DB_USER=chama
GREETING=Habari
```

`NAME` is never set, so `${NAME}` becomes nothing.

**Limits:** 2 seconds per test case.

**Go tip**

Read lines with a `bufio.Scanner`. To expand, walk the value with an index and build the result in a `strings.Builder`. When you see `${`, look for the next `}` with `strings.IndexByte(v[i+2:], '}')`. If it's `-1`, write the `$` and carry on.

Judge it from the repo root with `go run ./tools/kata 21`.
