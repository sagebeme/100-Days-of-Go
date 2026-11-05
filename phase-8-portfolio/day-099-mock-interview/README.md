# Day 99: Mock Interview

## The brief

You've applied for a junior backend role at Twende Mobility (a made-up Nairobi startup). Their recruiter, Faith, emails:

> Thanks for applying! The next step has two parts. First, a 90-minute online coding test: four problems plus one harder one, judged automatically. Then a 45-minute call with one of our engineers to design a system together. We care how you think as much as whether you finish, so talk us through it. Good luck!

Today you sit both, timed, as if they were real. Treat it as a rehearsal: the point is to find out where you stall under a clock while it costs nothing.

```
Part 1, coding (90 min):
  1. PairForBudget([350 120 500 230], 350)       →  1, 3, true            10 min
  2. MergeShifts([{360 600} {540 720} {900 1080}]) →  [{360 720} {900 1080}]  15 min
  3. LongestRun([T T F T T T F F T], 1)           →  6                     20 min
  4. FewestNotes([1 40 50], 80)                   →  2                     20 min
  5. The kata: Fewest Changes Fare                →  350                   25 min

Part 2, system design (45 min): Twende Seats, in starter/SYSTEM_DESIGN.md
```

**Acceptance criteria**

- Part 1: all tests in `starter/` pass, and `go run ./tools/kata 99` passes all cases, within 90 minutes in total. Record what you finished in time and what you finished after.
- Part 2: `starter/SYSTEM_DESIGN.md` filled in within 45 minutes, then self-marked with the rubric below.
- A short note to yourself: the two things that cost you the most time.

## What you'll use

Everything, under time pressure. The problems map to these days:

- Problem 1: hashing and sets, [Day 28](../../phase-3-algorithms/day-028-duplicate-receipts/).
- Problem 2: sorting with a custom order, [Day 31](../../phase-3-algorithms/day-031-sorting-showdown/).
- Problem 3: sliding windows, [Day 30](../../phase-3-algorithms/day-030-sliding-window/).
- Problem 4: dynamic programming, [Day 39](../../phase-3-algorithms/day-039-dynamic-programming/).
- The kata: graphs, [Day 37](../../phase-3-algorithms/day-037-graphs/).
- Part 2: [Day 77: System Design Basics](../../phase-7-production/day-077-system-design-basics/), the booking API from [Day 64](../../phase-5-backend/day-064-matatu-booking-api/), and retries from [Day 79](../../phase-7-production/day-079-background-jobs/).
- The format of [Day 40: Mock HackerRank Test](../../phase-3-algorithms/day-040-mock-hackerrank-test/), one level harder.

### Talking through a solution

In a live interview, silence reads as being stuck. A steady routine helps:

1. **Repeat the problem back** in your own words, with one example worked by hand. Half of all wrong answers solve a different problem.
2. **Ask about the edges**: empty input, duplicates, negative numbers, how big `n` gets. The size of `n` tells you the complexity they expect: 10^5 means O(n log n) or better.
3. **Say the brute force first** and its cost: "Checking every pair is O(n²), too slow for 10^5." Then improve it out loud. A correct slow answer beats a half-written fast one.
4. **Name the pattern** before typing: "This is a sliding window, because..." Interviewers mark the reasoning.
5. **Narrate while you code**, briefly: what this loop does and what's true at the top of it.
6. **Test by hand** with your example and one edge case before saying "done". Find your own bug before they do.
7. **State the complexity**, time and space, without being asked.

If you're stuck, say so and say what you're trying: "I think a map from value to index gets this to one pass; let me check it handles duplicates." Interviewers often give hints, but only to people who show where they are.

## Steps

**Part 1: the coding test (90 minutes)**

1. Set a timer for 90 minutes. Close the solution folders. Docs (`go doc`, pkg.go.dev) are allowed, as on HackerRank; AI assistants and searching for the problem are not.
2. Read all five problems first (the four doc comments in `starter/problems.go` and `kata/README.md`), then do them in the order that gets you the most points. Skip a problem after its time box and come back.
3. Check as you go:

```bash
go test ./phase-8-portfolio/day-099-mock-interview/starter
go test -run TestMergeShifts ./phase-8-portfolio/day-099-mock-interview/starter
go run ./tools/kata 99
```

4. When the timer goes, write down what passed. Then finish the rest without the clock, because you still need to know how.

**Part 2: system design (45 minutes)**

5. Set a timer for 45 minutes and fill in `starter/SYSTEM_DESIGN.md`. Speak your answer out loud as you write, as if someone were on the call. Better still, ask a friend to play the interviewer and push back.
6. Mark it with the rubric, honestly. Then read `solution/SYSTEM_DESIGN.md` and note what you missed.

### Self-marking rubric (20 points)

| Area | Points | Full marks if you... |
|---|---|---|
| Questions and requirements | 3 | asked about seat numbering, hold time and failure cases before designing; split functional from non-functional; said correctness beats availability for booking |
| Numbers | 2 | worked out bookings per second at peak (about 30) and concluded one database is enough |
| API | 2 | listed the main endpoints with methods and status codes, including `409` for a taken seat and idempotency keys |
| Data model | 3 | had trips, holds and payments, and a database constraint that makes double booking impossible |
| High-level design | 3 | drew clients, stateless API servers, the database, a queue with workers, and M-Pesa with its callback, and said what each does |
| Double booking | 2 | explained holds that expire, the unique constraint, and why an in-memory lock isn't enough |
| Payment callback | 2 | handled late, duplicate and missing callbacks: idempotency on the checkout ID, a status query, a rule for late payments |
| Failures and scaling | 2 | said what happens when the database, M-Pesa or SMS is down, and how December is handled |
| Trade-offs | 1 | named at least two choices and what you gave up for each |

16 or more is a strong answer for a junior role. Under 10: re-read [Day 77](../../phase-7-production/day-077-system-design-basics/) and try the question again in a week.

**Stretch goals**

- Do Part 1 again in a week, in 60 minutes.
- Record yourself solving one problem out loud on video. Watch it back. It's painful and it works.
- Swap system design questions with a friend from the course and interview each other.

## The kata: Fewest Changes Fare

Find the cheapest fare between two stages using at most `K` changes of matatu. Plain Dijkstra gives the wrong answer because it ignores the limit; Bellman-Ford stopped after `K + 1` rounds, or Dijkstra over (stage, rides) pairs, gets it right. It's the hardest problem in the set: LeetCode's "cheapest flights within K stops". Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 99
```

## Check yourself

1. The interviewer says `n` can be 10^5. What does that tell you about the complexity they expect, and why say the brute force out loud anyway?
2. Predict the output (run it against your solution):

   ```go
   fmt.Println(PairForBudget([]int{350, 120, 500, 230}, 350))
   fmt.Println(MergeShifts([]Shift{{900, 1080}, {360, 600}, {540, 720}, {720, 800}}))
   fmt.Println(LongestRun([]bool{true, true, false, true, true, true, false, false, true}, 0))
   fmt.Println(FewestNotes([]int{1, 40, 50}, 80), FewestNotes([]int{40, 50}, 30))
   ```

3. Spot the bug. `FewestNotes([]int{1, 40, 50}, 80)` returns 31 instead of 2:

   ```go
   sort.Sort(sort.Reverse(sort.IntSlice(denoms)))
   count := 0
   for _, d := range denoms {
       count += amount / d
       amount %= d
   }
   return count
   ```

   Why does this pass for Kenyan notes but not for `{1, 40, 50}`? (It also sorts the caller's slice.)
4. In the system design, why is an in-memory lock not enough to stop double booking, and what is?

<details>
<summary>Answers</summary>

1. O(n log n) or better: O(n²) is about 10^10 steps, far too slow. Saying the brute force first shows you understand the problem, gives a correct answer to fall back on, and sets up the improvement you explain next.
2. `1 3 true`, `[{360 800} {900 1080}]` (touching shifts merge too), `3`, and `2 -1` (40 + 40; 30 can't be made from 40s and 50s).
3. Greedy takes the biggest note first: 50, then 30 ones, 31 notes. It works for Kenyan notes because their values are chosen so greedy is always right; in general you need dynamic programming over every amount up to the target (Day 39). Work on a copy if you must sort.
4. A lock in one process doesn't stop another API server, and the servers are stateless and several. Holds that expire, and a unique constraint in the database on (trip, seat), make a second booking of the same seat impossible whichever server it reaches.

</details>

## When you're stuck

- **`PairForBudget([3 3 7], 10)` returns `1, 2`**: you overwrote the index of the first `5`. Only store a price's index the first time you see it.
- **`PairForBudget` pairs an item with itself**: look up `budget - p` before you add `p` to the map.
- **`MergeShifts` misses touching shifts**: compare with `<=`, not `<`. And keep the larger end: a shift inside another doesn't shorten it.
- **`MergeShifts` changes the caller's slice**: sorting sorts in place. `slices.Clone` first.
- **`LongestRun` is right but times out**: restarting the count from every week is O(n²). Move the left edge of the window forward only when there are too many misses.
- **`FewestNotes([1 40 50], 80)` gives 31**: greedy fails for this set. Build up the best answer for every amount from 0 to `amount`.
- **The kata gives 300 for the example**: one round of relaxation chained several rides. Copy the costs at the start of each round and read only from the copy.
- **You ran out of time on the design**: you spent it on the diagram. Get the numbers and the data model down early; they drive everything else.
- **Still stuck?** Open `solution/problems.go`, `kata/solution/main.go` and `solution/SYSTEM_DESIGN.md`, read them once, close them, then write your own versions from memory.

## Make it portfolio-ready

This day's portfolio piece is you, not a repository.

- **README**: keep a private `interview-log.md`: the date, what you finished in the time, the two things that cost you most, and your rubric score. Practise against it until the numbers improve.
- **Demo**: turn your design into a one-page write-up with a clean diagram, "How I'd build seat booking for a matatu SACCO", and post it. It shows how you think to people who'll never see your code.
- **Tests and CI**: if you keep solving practice problems, keep them in a repository with tests and a CI badge. A green history of practice is a quiet signal of discipline.
- **Deploy link**: not needed. A link to your best write-up does the same job.
- **What you'd do next**: book a real mock interview (a friend in the industry, a meetup, or a peer from the course), and do one timed problem set a week until your first offer.
