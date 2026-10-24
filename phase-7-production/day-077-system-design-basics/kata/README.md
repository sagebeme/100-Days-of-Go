# Kata: Hash Ring

Twende keeps seat maps in several cache servers. To decide which server holds a trip, it uses a **consistent hashing ring**: the numbers 0 to 4 294 967 295 wrap round like a clock face. Each server sits at one or more points on the ring (its "virtual nodes"). A key belongs to the first server point at or after the key's hash, going clockwise. A key past the last point wraps round to the first one. When a server is added or removed, only the keys next to its points move.

Replay a list of changes and lookups on the ring.

**Input**

- Line 1: `q`, the number of operations (1 ≤ q ≤ 10 000)
- Then `q` lines, each one of:
  - `add <name> <pos>`: put a point for server `name` at position `pos`
  - `remove <name>`: take every point of server `name` off the ring (nothing happens if it has none)
  - `get <hash>`: find the server for a key with this hash

Names are 1 to 20 letters, digits or hyphens. Positions and hashes are integers from 0 to 4 294 967 295. No two points on the ring are ever at the same position at the same time. At most 2 000 `add` lines appear.

**Output**

For each `get`, one line: the server's name, or `NONE` if the ring is empty.

**Example**

```
Input                   Output
8                       cache-b
add cache-a 100         cache-a
add cache-b 2000        cache-a
add cache-a 3000000000  NONE
get 150
get 4000000000
remove cache-b
get 150
remove cache-a
get 5
```

Hash 150 is after point 100 and before 2000, so it goes to `cache-b`. Hash 4 000 000 000 is past the last point, so it wraps to the point at 100. Once `cache-b` is gone, 150 moves on to the next point, 3 000 000 000, which is `cache-a`.

**Limits:** 2 seconds per test case. Positions need 64 bits; Go's `int` has them.

**Go tip:** keep the points in a slice sorted by position, with the server name beside each one. `sort.Search(len(ring), func(i int) bool { return ring[i].pos >= h })` finds the first point at or after `h` in O(log n); if it returns `len(ring)`, wrap to 0. Adding is `slices.Insert` at the index `sort.Search` gives you. Removing is `slices.DeleteFunc`.

Judge it from the repo root with `go run ./tools/kata 77`.
