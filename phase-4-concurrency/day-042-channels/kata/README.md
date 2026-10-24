# Kata: Round Robin

One M-Pesa agent is serving a queue of customers, and some of them have long jobs. To be fair, the agent works on the customer at the front for at most `q` minutes. If the customer's job is done, they leave. If not, they go to the back of the queue with whatever time is left. Then the agent serves the next customer.

Print the customers in the order they leave, each with the minute they leave at. The agent starts at minute 0 and never rests.

**Input**

- Line 1: `n` and `q`, the number of customers and the most the agent spends on one turn (1 ≤ n ≤ 100 000, 1 ≤ q ≤ 1 000)
- The next `n` lines: a name and the minutes that customer's job needs, in queue order (names are 1 to 10 letters and digits; 1 ≤ minutes ≤ 50 000; all the minutes add up to at most 1 000 000)

**Output**

`n` lines: the name and the finishing minute of each customer, in the order they finish.

**Example**

```
Input           Output
5 100           p2 180
p1 150          p5 400
p2 80           p1 450
p3 200          p3 550
p4 350          p4 800
p5 20
```

Minute by minute: `p1` gets 100 minutes and goes to the back with 50 left (minute 100). `p2` needs only 80 and leaves at 180. `p3` gets 100 and goes back with 100 left (280). `p4` gets 100 and goes back with 250 left (380). `p5` needs 20 and leaves at 400. `p1` finishes its last 50 at 450 and `p3` its last 100 at 550. Only `p4` is left: it takes two more full turns and then 50 minutes, leaving at 800.

**Limits:** 2 seconds per test case.

**Go tip**

A channel is a first-in, first-out queue, which is why this puzzle is here. In a program with one goroutine, though, a slice is the simpler queue: read the front with `queue[0]`, drop it with `queue = queue[1:]`, and add to the back with `append`. Keep each customer as a small struct with a name and the minutes left. Read names with `fmt.Fscan` into a `string`, and print with a `bufio.Writer` because the output can be long.

Judge it from the repo root with `go run ./tools/kata 42`.
