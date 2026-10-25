# Kata: Decode the SMS

Two cousins in Machakos text each other in a secret code: `A` is `1`, `B` is `2`, and so on up to `Z` = `26`. They send only the digits, with no spaces. So `226` could be `BZ` (2 26), `VF` (22 6) or `BBF` (2 2 6). Count how many different messages a string of digits could be. The count gets huge, so print it modulo 1 000 000 007.

A letter is `1` to `9` on its own, or two digits from `10` to `26`. A `0` can only be the second digit of `10` or `20`. `06` is not a letter.

**Input**

- Line 1: `n` (1 ≤ n ≤ 2 × 10^5)
- Line 2: `n` digits with no spaces

**Output**

One line: the number of ways to decode the digits, modulo 1 000 000 007. It can be 0.

**Example**

```
Input           Output
3               3
226
```

More cases: `12` gives 2 (`AB`, `L`), `06` gives 0, `100` gives 0 (there's no letter for `00` or `0`), `11106` gives 2.

**Limits:** 2 seconds per test case.

**Read the constraints.** Trying every split with plain recursion is exponential: a string of 2 × 10^5 ones has more decodings than atoms in the universe. But the number of ways to decode the rest from position `i` is the same every time you get there. Work it out once per position: `ways[i] = ways[i+1]` if `s[i]` is a letter on its own, plus `ways[i+2]` if `s[i:i+2]` is a letter from 10 to 26. That's O(n).

**Go tip**

Fill the table from the end, with `ways[n] = 1` for the empty rest. You only ever look one and two places ahead, so two variables can replace the slice. Take the modulo after every addition.

Judge it from the repo root with `go run ./tools/kata 39`.
