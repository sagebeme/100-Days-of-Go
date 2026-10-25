# Kata: Shared Refresh

The Twende page logs in at time 0 and gets an access token that works until time `A`. After that the server answers 401, and the page must refresh the token, which takes `D` seconds. A good client shares one refresh: requests that come in while a refresh is running wait for it instead of starting their own.

Process the requests in order. For a request at time `t`:

1. If a refresh is running (`t` is before the moment it finishes), the request waits and is sent when the refresh finishes.
2. Otherwise, if the token still works (`t < expiry`), the request is sent at `t`.
3. Otherwise the token has expired (`t >= expiry`). Start a refresh at `t`. It finishes at `t + D`, the request is sent then, and the new token works until `t + D + A`.

Print when each request is actually sent, and how many refreshes there were.

**Input**

- Line 1: `n A D` (1 ≤ n ≤ 10 000; 1 ≤ A, D ≤ 10^6)
- Line 2: `n` request times, in non-decreasing order (0 ≤ t ≤ 10^9)

**Output**

- `n` lines: the time each request is sent
- Then `refreshes R`

**Example**

```
Input                    Output
6 60 2                   10
10 59 60 61 62 130       59
                         62
                         62
                         62
                         132
                         refreshes 2
```

At 60 the token has expired, so a refresh runs from 60 to 62. The requests at 60 and 61 wait for it, and the one at 62 goes straight out with the new token, which works until 122. At 130 it has expired again.

**Limits:** 2 seconds per test case.

**Go tip:** three variables carry the whole state: `expiry`, `refreshDone` (start it at -1) and a counter. Print with a `bufio.Writer` and `defer out.Flush()`, since there are thousands of lines.

Judge it from the repo root with `go run ./tools/kata 70`.
