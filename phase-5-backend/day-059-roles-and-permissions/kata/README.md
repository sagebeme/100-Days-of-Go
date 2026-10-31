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

**Go tip:** for each role, walk its parents once with a slice used as a queue and a `seen` set (a breadth-first search, Day 37), and keep the roles you reach. Then don't check those roles one by one for every question: 200 000 questions times up to 100 roles is 20 million map lookups, too slow for 2 seconds. Turn it around instead. Give each role a number from 0 to 99, and store a set of roles as two `uint64`s, one bit per role (`[2]uint64`). Keep, for each action, the set of roles granted it, and for each role, the set of roles it reaches. A question is then one map lookup and an AND of two sets: `reach[r][0]&granted[0] != 0 || reach[r][1]&granted[1] != 0`.

Judge it from the repo root with `go run ./tools/kata 59`.
