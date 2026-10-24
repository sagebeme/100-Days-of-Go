# Kata: Permission Matrix

Twende's permissions now live in a config file instead of code. Roles get actions, roles can inherit everything another role has, and users are given a role. Answer a list of "may this user do this?" questions.

Each rule is one of:

- `GRANT role action`: the role may perform the action.
- `INHERIT child parent`: the child role also gets every permission the parent has, including what the parent inherits. Chains can be long, and they can loop (`a` inherits `b`, `b` inherits `a`); a loop just means those roles share everything.
- `USER name role`: the user has this role. If a user appears twice, the later line wins.

Read every rule before answering. A user with no `USER` line, or an action nobody was granted, is denied. Names are case-sensitive.

**Input**

- Line 1: `n`, the number of rules (1 ≤ n ≤ 200 000)
- Next `n` lines: one rule each
- Then a line with `q`, the number of questions (1 ≤ q ≤ 200 000)
- Next `q` lines: `name action`

There are at most 100 distinct roles. Names have no spaces.

**Output**

`q` lines: `ALLOW` or `DENY`.

**Example**

```
Input                              Output
9
GRANT passenger trips:list
GRANT passenger bookings:create
GRANT conductor bookings:checkin
INHERIT conductor passenger
INHERIT admin conductor
GRANT admin trips:create
USER wanjiku passenger
USER kamau conductor
USER amina admin
6
wanjiku trips:list                 ALLOW
wanjiku bookings:checkin           DENY
kamau bookings:create              ALLOW
kamau trips:create                 DENY
amina bookings:create              ALLOW
otieno trips:list                  DENY
```

Amina is an admin, admins inherit from conductors, and conductors inherit from passengers, so she may create bookings.

**Limits:** 2 seconds per test case.

**Go tip:** keep `map[string]map[string]bool` for grants and `map[string][]string` for parents. For each role, walk its parents once with a slice used as a queue and a `seen` map (a breadth-first search, Day 37), and cache the list of roles you reach. A map of maps returns `false` for a missing inner key even when the outer key is missing too: `grants["nobody"]["x"]` is safe to read.

Judge it from the repo root with `go run ./tools/kata 59`.
