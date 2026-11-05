# Day 65: TypeScript for Go Developers

## The brief

Phase 5 built a matatu booking API in Go. Phase 6 builds the website that talks to it, and websites are written in TypeScript. You already know most of it: TypeScript is JavaScript with a type checker, and many of its ideas have a Go twin. Today you write five small functions about Nairobi routes and fares, with no browser yet, so you can meet the language on its own.

```
describePayment({ kind: "mpesa", phone: "0712345678", amountKsh: 100 })  →  "Ksh100 by M-Pesa from 0712345678"
describePayment({ kind: "cash", amountKsh: 150 })                         →  "Ksh150 in cash"
fareFor(ROUTES, "58")                                                     →  150
fareFor(ROUTES, "999")                                                    →  undefined
groupBy(ROUTES, (r) => r.boardsAt).get("Railways")                        →  [route 111, route 125]
parseRoute(JSON.parse('{"number":46}'))                                   →  throws Error("invalid route: ...")
await totalFare(["46K", "58", "125"], lookup)                             →  350
```

The routes are real (checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026). Routes change, so confirm with the conductor.

## What you'll use

Set up once: `cd phase-6-frontend && npm ci` (see [the phase README](../README.md)).

| Go | TypeScript | Note |
|---|---|---|
| `type Route struct { Fare int }` | `interface Route { fareKsh: number }` | a TS interface is a shape. Any object with those fields fits; there is no `implements` needed, like Go interfaces |
| `int`, `float64` | `number` | one number type, a 64-bit float. Whole numbers are exact up to 2^53 |
| `x, ok := m[k]` | `const x: number \| undefined = ...` | "maybe missing" is part of the type. `?.` and `??` handle it |
| `switch v := p.(type)` | `switch (p.kind)` | a **discriminated union**: each shape has a `kind` field with a fixed string, and TS narrows `p` inside each `case` |
| `func Map[T any](...)` | `function groupBy<T, K>(...)` | generics, with the type parameters in `<>` |
| `any` from `json.Unmarshal` | `unknown` | `unknown` forces you to check before use. Avoid `any`: it turns the checker off |
| goroutines + `sync.WaitGroup` | `await Promise.all([...])` | one thread, but many requests can wait at once. `await` pauses this function, not the program |
| `return x, err` | `throw new Error(...)` | errors are thrown; an `async` function's throw becomes a rejected promise |

```ts
type Payment = { kind: "cash"; amountKsh: number } | { kind: "mpesa"; phone: string; amountKsh: number };

function who(p: Payment): string {
  if (p.kind === "mpesa") return p.phone; // narrowed: p.phone exists here
  return "cash";                          // and here p can only be the cash shape
}

const fare = routes.find((r) => r.number === "58")?.fareKsh; // number | undefined
const fares = await Promise.all(["46K", "58"].map(lookup));   // starts both, then waits for both
```

TypeScript types disappear when the code runs. That's why `parseRoute` exists: data from the network can be anything, whatever your types say, so check it with `typeof` the way Go's decoder checks JSON for you.

## Steps

1. Open `starter/routes.ts`. Keep every name and signature; the tests in `starter/routes.test.ts` import them.
2. `describePayment`: `switch (p.kind)` with a `case` for each kind. Use a template string: `` `Ksh${p.amountKsh} in cash` ``. Add a `default` that assigns `p` to a `never` variable, so a new kind that you forget becomes a compile error.
3. `fareFor`: `routes.find(...)` returns the route or `undefined`. `?.fareKsh` reads the fare only if there was one.
4. `groupBy`: loop with `for (const item of items)`, keep a `Map<K, T[]>`, push into the list or `set` a new one. A `Map` remembers insertion order, so the keys come out in the order you met them.
5. `parseRoute`: reject anything that isn't an object (remember `typeof null === "object"`), then check each field with `typeof`. Throw `new Error("invalid route: ...")`.
6. `totalFare`: `numbers.map(lookup)` starts every lookup, `await Promise.all(...)` waits for all of them, and `reduce` adds them up. Awaiting inside a `for` loop would run them one after another and fail the second test.
7. Run `npm run typecheck`, then the tests, from `phase-6-frontend`:

```bash
npm test -- day-065
```

There's no page to look at today. From [Day 66](../day-066-first-react-app/), `npm run dev` shows your work in the browser.

## The kata: Payment Events

A till log where the first word of each line decides what follows: the same union, read in Go. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 65
```

## Check yourself

1. TypeScript types disappear when the code runs. Why does that mean `parseRoute` has to check every field with `typeof`, even though `fetch` results are "typed" as `Route[]` in your code?
2. Predict the output (run it with `node -e` if you like):

   ```ts
   console.log(typeof null);
   console.log(0 ?? "Ask the conductor");
   console.log(0 || "Ask the conductor");
   console.log([{ number: "58", fareKsh: 150 }].find((r) => r.number === "999")?.fareKsh);
   ```

3. Spot the bug. With three lookups that each take 100 ms, the second test fails because this takes 300 ms:

   ```ts
   async function totalFare(numbers: string[], lookup: (n: string) => Promise<number>) {
     let total = 0;
     for (const n of numbers) total += await lookup(n);
     return total;
   }
   ```

4. What does the `default` branch that assigns `p` to a `never` variable buy you in `describePayment`?

<details>
<summary>Answers</summary>

1. The checker only sees your source code. At run time the server can send anything, and nothing compares it with your types. A wrong field turns into `undefined` far from where it came in. Checking with `typeof` at the edge, like Go's decoder, turns bad data into one clear error.
2. `object` (a long-standing JavaScript quirk, which is why `parseRoute` checks for `null` itself), `0`, `Ask the conductor`, `undefined`. `??` only replaces `null` and `undefined`, while `||` replaces any falsy value, including a real Ksh0.
3. `await` inside the loop waits for each lookup before starting the next one. Start them all, then wait once: `(await Promise.all(numbers.map(lookup))).reduce((a, b) => a + b, 0)`.
4. Exhaustiveness. Inside `default`, TypeScript has ruled out every known kind, so `p` has type `never`. If someone adds `{ kind: "card" }` to the union and forgets a `case`, `p` is no longer `never` and the assignment becomes a compile error, not a wrong answer at run time.

</details>

## When you're stuck

- **`'p.phone' does not exist on type 'Payment'`**: you read a field before narrowing. Check `p.kind` first; inside that branch the field exists.
- **`Type 'Payment' is not assignable to type 'never'`** in your `default`: a `case` is missing. That error is the point of the `never` trick.
- **`'value' is of type 'unknown'`**: check it with `typeof` before reading fields, then read through `value as Record<string, unknown>`.
- **`parseRoute(null)` doesn't throw**: `typeof null` is `"object"`. Test `value === null` too.
- **"lookups started before the first one finished" fails**: you awaited inside the loop. Start them all with `map`, then await once.
- **`... is a type and must be imported using a type-only import`**: the config has `verbatimModuleSyntax` on. Write `import { type Route } from "./routes"`.
- **Still stuck?** Open `solution/routes.ts`, read it once, close it, then write your own version from memory.
