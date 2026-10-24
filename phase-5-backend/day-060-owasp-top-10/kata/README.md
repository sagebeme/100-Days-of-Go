# Kata: Blocked Addresses

Twende's webhook sender must never connect to an address inside its own network. Given a list of blocked ranges in CIDR notation and a list of IPv4 addresses, print whether each address is blocked.

`10.0.0.0/8` means "every address whose first 8 bits match 10.0.0.0", so 10.0.0.0 to 10.255.255.255. The part before the `/` may have bits set after the prefix (`192.168.1.77/24`); ignore them, so it means `192.168.1.0/24`. A `/0` block matches everything and a `/32` block matches one address.

An address is valid only if it has exactly four parts separated by dots, and each part is 1 to 3 digits, at most 255, with no leading zero (`0` is fine, `01` is not). Leading zeros are refused because some software reads `010` as octal 8, a classic way to slip past a filter.

**Input**

- Line 1: `m`, the number of blocks (1 ≤ m ≤ 200)
- Next `m` lines: a block, `a.b.c.d/n` with 0 ≤ n ≤ 32. Blocks are always well formed.
- Then a line with `q` (1 ≤ q ≤ 100 000)
- Next `q` lines: an address to check. It has no spaces but may be invalid.

**Output**

`q` lines: `BLOCK` if the address is in any block, `ALLOW` if it's valid and in none, `INVALID` if it isn't a valid address.

**Example**

```
Input              Output
3
127.0.0.0/8
10.0.0.0/8
169.254.0.0/16
4
127.0.0.1          BLOCK
196.201.214.200    ALLOW
169.254.169.254    BLOCK
256.1.1.1          INVALID
```

**Limits:** 2 seconds per test case.

**Go tip:** turn each address into a `uint32`: `v = v<<8 | part` for each of the four parts. A `/n` mask is `^uint32(0) << (32 - n)`. For `/0` that shifts by 32, which gives 0 in Go, exactly the mask you want. (In C it's undefined, and Java gives back the original value, so other languages need a special case.) Then `ip & mask == network & mask` is the whole test. In real code, `netip.ParsePrefix` and `Prefix.Contains` do this for you.

Judge it from the repo root with `go run ./tools/kata 60`.
