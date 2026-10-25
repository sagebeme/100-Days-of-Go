# Day 82: Threat Modelling

## The brief

Twende is about to take real money through M-Pesa. Before launch, the SACCO's insurer asks for a threat model: "What can go wrong, what are you doing about it, and how do you know you thought of everything?" You can't promise everything, but you can be systematic. Today you describe Twende as boxes and arrows in code, generate the STRIDE questions for every box and arrow so nothing is skipped, and rank the answers so the team fixes the worst first.

```
STRIDE(TwendeModel())  →  42 threats, for example
  Passenger app: Spoofing                  Could someone pretend to be Passenger app?
  Booking API: Elevation of privilege       Could someone use Booking API to do more than they're allowed?
  M-Pesa -> Booking API: Spoofing           Could someone pretend to be M-Pesa -> Booking API?   (crosses a boundary, unauthenticated)

Score(4, 5)  →  20    Level(20)  →  "critical"
Prioritise(risks)  →  the fake M-Pesa callback first, the noisy-but-harmless ones last
```

## What you'll use

**A data-flow diagram.** Four kinds of thing:

- **External entities**: people and systems you don't run. Passengers, M-Pesa.
- **Processes**: code you run. The booking API, the seat service.
- **Data stores**: where data rests. Postgres, Redis.
- **Data flows**: arrows between them, labelled with what moves.

Draw **trust boundaries** where the level of trust changes: between the internet and your API, between your API and a partner, between the DMZ and the internal network. Most attacks happen on arrows that cross a boundary. In code, each component has a `Zone`, and a flow between different zones crosses a boundary.

```
 [Passenger app]──login, bookings──▶(Booking API)──BookSeat──▶(Seat service)──▶[[Postgres]]
     internet          │   ▲ dmz        │              internal
                 STK push  callback     └──cached trips──▶[[Redis]]
                       ▼   │
                     [M-Pesa]  partner
```

**STRIDE**, one question per letter:

| Letter | Threat | Property it breaks | Typical fix |
|---|---|---|---|
| S | **Spoofing**: pretending to be someone else | authentication | passwords, tokens, mTLS, signed callbacks |
| T | **Tampering**: changing data or code | integrity | TLS, signatures, parameterised SQL, access control |
| R | **Repudiation**: "I never did that" | non-repudiation | audit logs with who, what, when |
| I | **Information disclosure**: data reaching the wrong people | confidentiality | encryption, authorisation, no secrets in logs |
| D | **Denial of service**: stopping it working | availability | rate limits, timeouts, body size limits |
| E | **Elevation of privilege**: doing more than allowed | authorisation | deny by default, object-level checks (Day 59) |

**STRIDE per element.** Not every letter fits every box. An external entity can be spoofed and can deny things, but you can't patch its code. A process can suffer all six. A data store can be tampered with, leak or fill up. A flow can be tampered with, read or cut. `PerKind` and `PerFlow` hold that table, and today's code adds one rule: a flow that crosses a boundary without authentication can also be **spoofed**.

**Rating risks.** For each threat, the team agrees a **likelihood** and an **impact**, each from 1 to 5. Risk = likelihood × impact. Ties go to the higher impact: a rare disaster beats a frequent annoyance. This is a rough tool for ordering a to-do list, not a precise measurement.

**Walking through Twende.** A few of the 42, with the answers a team might write:

- **M-Pesa → Booking API, Spoofing** (likelihood 4, impact 5, critical). The payment callback is just an HTTPS POST to a URL. Anyone who finds that URL can post "payment received" and ride free. Fix: never trust a callback on its own. Before marking a booking paid, confirm the transaction with M-Pesa's own API, check the amount and phone match the booking, accept each receipt number once, and keep the callback URL unguessable and restricted to the addresses Safaricom publishes.
- **Booking API, Elevation of privilege** (3, 5, high). A passenger reads someone else's booking by changing the ID. Fix: the object-level check from Day 59, and a test grid of every route against every role.
- **Passenger app, Repudiation** (3, 3, medium). "I never booked that seat." Fix: an audit log of who booked what and when, with the request ID (Day 81), kept where the API can append but not edit.
- **Booking API → Seat service, Spoofing and Tampering** (2, 4, medium). The gRPC call (Day 80) has no TLS and no authentication, so anything inside the network can book seats. Fix: mutual TLS between services, or at least a service token.
- **Booking API, Denial of service** (4, 3, high). A script floods `/bookings`. Fix: rate limiting (Day 47), timeouts and body limits (Day 84).
- **Redis, Information disclosure** (2, 2, low). Cached trips aren't secret. Note it, and move on.

Every row in [owasp-checklist.md](../../reference/owasp-checklist.md) maps to a STRIDE letter: A01 broken access control is E, A02 cryptographic failures is I, A03 injection is T, A07 authentication failures is S, and A09 logging failures is R. Use the checklist for the fixes once STRIDE has found the problems.

## Why it matters at work

Threat models are how a design review asks "is this safe?" before the code exists, when changes are cheap. Many teams require one for anything that touches money, personal data or a new external partner, and auditors and insurers ask to see it. The value isn't the document; it's the hour where someone says "wait, who can call the callback URL?". Keeping the model as code means it's reviewed in pull requests like everything else: a new arrow in the diagram shows up as a diff, and a test can insist that every unauthenticated boundary crossing has a risk rating and an owner.

## Steps

1. Open `starter/model.go`. The types, the STRIDE tables and the questions are given. `starter/twende.go` is the Twende model.
2. Write `Validate`. Use a `map[string]bool` of names seen: a repeat is `ErrDuplicate`, a kind missing from `PerKind` is `ErrBadKind`, and a flow end not in the map is `ErrUnknown`. Wrap with `fmt.Errorf("%w: %q", ErrX, name)` so the message names the culprit (Day 17).
3. Write `zone` and `CrossesBoundary`.
4. Write `STRIDE`. Validate first. Then one `Threat` per category in `PerKind[c.Kind]` for each component, then one per category for each flow, with `Spoofing` first when the flow crosses a boundary and isn't authenticated. A flow's target is `From + " -> " + To`. Set `Boundary` on flow threats.
5. Write `Score` and `Level`.
6. Write `Prioritise`. Check every rating first. Sort a `slices.Clone` with `slices.SortStableFunc`: score, then impact, both biggest first, then `slices.Index(Order, category)`, then `cmp.Compare` on the target.
7. Run the tests from the repo root:

```bash
go test ./phase-7-production/day-082-threat-modelling/starter
```

8. Now do it for real. Add a component to `TwendeModel` that you know Twende will need, such as an admin dashboard in an `office` zone, run `STRIDE`, and write likelihood, impact and a fix for its five worst threats.

## The kata: Attack Paths

Treat the threat model as a graph and find how many hops each sensitive asset is from the internet, with a breadth-first search from several starting points. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 82
```

## When you're stuck

- **`error "duplicate component" doesn't name "API"`**: wrap with `%w` and add the name: `fmt.Errorf("%w: %q", ErrDuplicate, c.Name)`.
- **`STRIDE(Twende) gave 40 threats, want 42`**: the two unauthenticated boundary crossings need their extra Spoofing threat.
- **`Spoofing flagged for flows ... want only the 2`**: Spoofing is added only when the flow crosses a boundary **and** isn't authenticated.
- **`Boundary = false, want true`**: you set `Boundary` only on the Spoofing threat. Set it on every threat for that flow.
- **The order of a flow's threats is wrong, or `PerFlow` changes between tests**: you appended to `PerFlow` itself. Build a new slice: `append([]Category{Spoofing}, PerFlow...)`.
- **`Prioritise order` puts Postgres before Redis**: on equal score and impact, STRIDE order (Tampering before Information disclosure) comes before the name.
- **`Prioritise changed its input`**: sort `slices.Clone(risks)`.
- **Still stuck?** Open `solution/model.go`, read it once, close it, then write your own version from memory.
