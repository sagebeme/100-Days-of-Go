# Kata: Delivery Count

Replay a chat server's event log and say how many people each message reached. It's the hub's room bookkeeping without the sockets: a map from room to a set of users.

**Input**

- Line 1: `n` (1 ≤ n ≤ 200 000)
- The next `n` lines: one event, `OP user room`, where `OP` is one of:
  - `JOIN`: the user joins the room. Joining a room you're already in does nothing.
  - `LEAVE`: the user leaves the room. Leaving a room you're not in does nothing.
  - `SAY`: the user sends a message to the room.
- Names are 1 to 20 characters with no spaces. A user can be in many rooms at once.

**Output**

- One line per `SAY`: how many *other* users in the room received it, or `denied` if the sender isn't in that room.
- A last line: the total number of deliveries.

**Example**

```
Input                    Output
8                        1
JOIN otieno route-125    0
JOIN wanjiku route-125   denied
SAY otieno route-125     0
JOIN kamau route-46K     1
SAY kamau route-46K
SAY kamau route-125
LEAVE wanjiku route-125
SAY otieno route-125
```

Otieno's first message reaches Wanjiku. Kamau is alone in `route-46K`, and he isn't in `route-125`, so he's denied there. After Wanjiku leaves, Otieno's message reaches nobody. The total is 1.

**Limits:** 2 seconds per test case.

**Go tip**

A set is a map to `struct{}`. Make the inner map the first time a room is used:

```go
rooms := map[string]map[string]struct{}{}
if rooms[room] == nil {
	rooms[room] = map[string]struct{}{}
}
rooms[room][user] = struct{}{}
```

Judge it from the repo root with `go run ./tools/kata 87`.
