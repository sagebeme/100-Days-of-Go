# Kata: LRU Cache

Twende's phone app keeps the last few seat maps it fetched, so going back a screen doesn't hit the network. It has room for only `c` of them. When it's full and a new one arrives, it throws out the one **least recently used**: the one nobody has read or written for the longest time. Both `get` and `put` count as using a key.

This is one of the most asked interview questions there is. The catch: every operation must be O(1), not a scan of the whole cache.

**Input**

- Line 1: `c q`, the capacity and the number of operations (1 ≤ c ≤ 100 000, 1 ≤ q ≤ 200 000)
- Then `q` lines, each one of:
  - `put <key> <value>`: store the value under the key, replacing any old value. If that adds a new key to a full cache, first remove the least recently used key.
  - `get <key>`: look the key up

Keys are integers from 1 to 1 000 000 000. Values are integers from 0 to 1 000 000 000.

**Output**

For each `get`, one line: the value, or `-1` if the key isn't in the cache.

**Example**

```
Input          Output
2 7            10
put 1 10       -1
put 2 20       30
get 1          10
put 3 30
get 2
get 3
get 1
```

`get 1` makes key 1 the most recently used, so when key 3 arrives the cache throws out key 2.

**Limits:** 2 seconds per test case.

**Go tip:** pair a map with a doubly linked list. `container/list` from the standard library is one: `l.PushFront(v)` returns an `*list.Element`, `l.MoveToFront(e)` and `l.Remove(e)` are O(1), and `l.Back()` is the least recently used. Keep `map[int]*list.Element`, and store the key inside the element's value too, so that when you evict `l.Back()` you know which map entry to delete. Read with `bufio` and write with a `bufio.Writer`; 200 000 lines is a lot for `fmt.Println`.

Judge it from the repo root with `go run ./tools/kata 78`.
