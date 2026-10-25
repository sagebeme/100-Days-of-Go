# Kata: Undo Button

The stage queue app needs an undo button, because conductors tap fast. Replay a list of actions and print the queue at the end.

- `add NAME`: put NAME at the end of the queue.
- `remove NAME`: take out the first NAME in the queue (the one nearest the front). If there's no NAME, nothing happens and the action isn't remembered.
- `undo`: reverse the most recent `add` or `remove` that hasn't been undone yet. An undone `remove` puts the passenger back exactly where they were. With nothing left to undo, nothing happens.

**Input**

- Line 1: `n`, the number of actions (1 ≤ n ≤ 5 000)
- Next `n` lines: one action each. Names are single words.

**Output**

One line: the queue from front to back, separated by spaces, or `empty`.

**Example**

```
Input             Output
6                 Amina Baraka Kamau Chebet
add Amina
add Baraka
add Kamau
remove Baraka
undo
add Chebet
```

The `undo` puts Baraka back between Amina and Kamau.

**Limits:** 2 seconds per test case.

**Go tip:** keep a history stack of what each action did, as a small struct: `type change struct { added bool; index int; name string }`. Undo pops the last one and does the opposite. To insert at `i`, use `slices.Insert(queue, i, name)`; to delete, `slices.Delete(queue, i, i+1)`.

Judge it from the repo root with `go run ./tools/kata 67`.
