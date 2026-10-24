# Kata: Time Travel Store

Your key-value store keeps every change in its log, so it could answer a better question than "what is the value now?". It could answer "what was the value at time `t`?". The chama treasurer wants exactly that: "What was the float on the 15th, before Otieno's withdrawal?"

Process a list of commands. `SET` and `DEL` change a key at a time. `GET` asks for the value a key had at a time: the value from the latest `SET` at or before that time, unless a `DEL` came after that `SET` and at or before the time.

**Input**

- Line 1: `Q`, the number of commands (1 ≤ Q ≤ 200 000)
- The next `Q` lines, one of:
  - `SET key value t`: from time `t`, `key` holds `value`
  - `DEL key t`: from time `t`, `key` has no value
  - `GET key t`: ask for the value of `key` at time `t`
- Keys and values are 1 to 10 letters and digits. Times are integers, 1 ≤ t ≤ 10^9.
- The times on `SET` and `DEL` lines strictly increase down the input. `GET` times can be anything.

**Output**

For each `GET`, one line: the value, or `-` if the key had no value at that time.

**Example**

```
Input                  Output
11                     4500
SET float 4500 10      -
SET till 247247 20     3200
GET float 15           4500
GET float 5            -
SET float 3200 30      3200
GET float 30           247247
GET float 29
DEL float 40
GET float 45
GET float 35
GET till 1000
```

At 15 the float is `4500`. At 5 nothing was set yet. At 30 the new `SET` counts, because "at or before" includes 30 itself. At 45 the float was deleted at 40. At 35 it was still `3200`.

**Limits:** 2 seconds per test case.

**Go tip**

Keep a slice of versions per key. Because the times only grow, `append` keeps each slice sorted, and `sort.Search` finds the answer in O(log n):

```go
// index of the first version after t, minus one: the latest version at or before t
i := sort.Search(len(h), func(i int) bool { return h[i].t > t }) - 1
```

Use `bufio.NewReader` and `bufio.NewWriter`, as on Day 27. With 200 000 lines, unbuffered output is the slow part.

Judge it from the repo root with `go run ./tools/kata 93`.
