# Kata: Account Lockout

Twende's login page locks an account after three wrong passwords in a row, so nobody can guess a passenger's password by trying all night. Replay a log of login attempts and print what the server answered to each one.

The rules, per phone number:

- An attempt is either `OK` (right password) or `FAIL` (wrong password).
- `OK` prints `OK` and resets the count of failures to 0.
- `FAIL` adds 1 to the count. If the count is now 3, print `FAIL LOCKED`: the account is locked for 300 seconds from this attempt, and the count goes back to 0. Otherwise print `FAIL`.
- While an account is locked (an attempt at time `t` with `t < lockTime + 300`), print `LOCKED`. The attempt isn't checked, so it doesn't count as a failure or a success.
- Each phone number has its own count and its own lock.

**Input**

- Line 1: `n`, the number of attempts (1 ≤ n ≤ 200 000)
- Next `n` lines: `t phone result`, where `t` is the time in seconds (0 ≤ t ≤ 10^9, never decreasing), `phone` is a string of digits and `result` is `OK` or `FAIL`

**Output**

`n` lines: `OK`, `FAIL`, `FAIL LOCKED` or `LOCKED`.

**Example**

```
Input                    Output
8
10 0712345678 FAIL       FAIL
20 0712345678 FAIL       FAIL
25 0722000000 OK         OK
30 0712345678 FAIL       FAIL LOCKED
100 0712345678 OK        LOCKED
329 0712345678 OK        LOCKED
330 0712345678 OK        OK
331 0712345678 FAIL      FAIL
```

The lock starts at 30 and lasts until 330, so the attempt at 329 is still locked and the one at 330 isn't.

**Limits:** 2 seconds per test case.

**Go tip:** keep a `map[string]*account` where `account` holds the failure count and the time the lock ends. A pointer lets you change the struct in place. With up to 200 000 lines of output, print through a `bufio.Writer` and `defer out.Flush()`, or the many small writes add up.

Judge it from the repo root with `go run ./tools/kata 57`.
