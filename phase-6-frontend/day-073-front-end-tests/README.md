# Day 73: Front-End Tests

## The brief

Every day so far came with its tests already written. Today you build the testing tools yourself. The component is finished: `TripList` shows the next departures from Railways and Ambassadeur, with a filter for trips that still have seats. Your job is a **fake `fetch`** that answers by method and path and records every call, so tests can talk to "the API" with no server. Then read the component tests that use it, break the component on purpose to watch them catch it, and meet Playwright, which drives a real browser.

```
const fake = createFakeFetch({ "GET /api/trips": () => json(200, TRIPS) });
await fake.fetch("http://localhost:8080/api/trips?route=58")   →  200, the trips
await fake.fetch("http://localhost:8080/api/nope", { method: "DELETE" })
                                                                →  404 {"error":"no fake for DELETE /api/nope"}
fake.calls                                                      →  ["GET /api/trips", "DELETE /api/nope"]

describeTrip(trip58, NOW)   →  "58 to Buruburu, leaves in 12 min, 3 seats left, Ksh150"
```

The trips are made up. Routes change, so confirm with the conductor.

## What you'll use

**Vitest** runs the tests (`describe`, `it`, `expect`), with `jsdom` standing in for the browser. **Testing Library** renders a component and finds things the way a user does: `getByRole`, `findByRole` (waits), `queryByRole` (gives `null` instead of throwing). **user-event** types and clicks like a person, with focus and keyboard events, which is why its calls are `await`ed.

**Test doubles.** Swap the slow, flaky or costly part for a fake you control. `vi.stubGlobal("fetch", fake.fetch)` replaces `fetch` for one test, and `vi.unstubAllGlobals()` in `afterEach` puts it back. A fake that routes by `"METHOD /path"` reads better than a chain of `mockResolvedValueOnce`, and its 404 tells you exactly which request you forgot to fake.

**The clock is an input.** "Leaves in 12 min" depends on now. `vi.useFakeTimers({ toFake: ["Date"] })` with `vi.setSystemTime(...)` freezes `Date.now()` while leaving real timers alone, so user-event and promises still work.

**Playwright** opens a real browser (Chromium, Firefox or WebKit) on the real dev server. `page.route` answers network calls inside the browser, so it still needs no Go server. It catches what jsdom can't: real CSS and layout, real focus, real navigation. It's slower and needs a browser download, so verify doesn't run it. A healthy suite has many fast unit and component tests and a few end-to-end tests for the journeys that make money.

```ts
await page.route("**/api/trips", (route) => route.fulfill({ json: trips }));
await page.goto("/");
await page.getByRole("checkbox", { name: "Only trips with seats" }).check();
await expect(page.getByRole("listitem")).toHaveCount(1);   // retries until true or timeout
```

## Why it matters at work

Front-end tests that break whenever markup changes get deleted; tests that query by role and text survive refactors and double as accessibility checks. Reviewers will ask for a test with every bug fix, will reject `setTimeout` sleeps in tests, and will push back on mocking a component's own internals instead of its edges (the network, the clock). A flaky end-to-end test that blocks deploys is a production incident of its own.

## Steps

1. In `starter/fakeFetch.ts`, write `json`: `new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })`.
2. Write `createFakeFetch`. Inside, `fakeFetch(input, init)` builds `new Request(input, init)`, makes the key from `req.method` and `new URL(req.url).pathname`, pushes it to `calls`, then returns the handler's answer or the 404. Return `{ fetch: fakeFetch, calls }`.
3. Run the tests. `fakeFetch.test.ts` checks your helper; `TripList.test.tsx` uses it to test the component.

```bash
npm test -- day-073
```

4. Prove the tests test something. In `starter/TripList.tsx`, change `t.seatsLeft > 0` to `t.seatsLeft >= 0`, run the tests, and watch "hides full trips when asked" fail. Put it back. Try another: delete the `!res.ok` check.
5. Write one test of your own in `TripList.test.tsx`: when every trip is full and the filter is ticked, the list has no items. Run it, then break the filter again to see your test fail.
6. Run the end-to-end spec in `e2e/trips.spec.ts`, from `phase-6-frontend`. The first command downloads a browser once (about 120 MB):

```bash
npx playwright install chromium
npx playwright test --config day-073-front-end-tests/playwright.config.ts
```

It starts `npm run dev` for you, on a simulated phone screen. Add `--headed` to watch it, or `--ui` to step through. To see the page yourself: `npm run dev -- day-073-front-end-tests/starter` (with no API running, you'll see the error state).

## The kata: Flaky Finder

Go through a week of CI results and report which tests are flaky and which are broken. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 73
```

## When you're stuck

- **`TypeError: Failed to parse URL from /api/trips`**: `Request` needs an absolute URL outside a browser page. The component uses `API_URL`, so check you're passing `input` through unchanged.
- **`calls` has `GET /api/trips?route=58`**: use `new URL(req.url).pathname`, which leaves out the query string.
- **`SyntaxError: Unexpected end of JSON input`**: `json` built a response with no body. `JSON.stringify` the body.
- **`Unable to find role="list"` right after render**: the list appears after the fetch resolves. Use `await screen.findByRole(...)`, not `getByRole`.
- **"leaves in" shows a huge number**: the fake clock isn't on. `vi.useFakeTimers({ toFake: ["Date"] })` and `vi.setSystemTime(NOW)` must run before `render`.
- **Playwright: `Executable doesn't exist`**: run `npx playwright install chromium` first. **`Timed out waiting ... from config.webServer`**: port 5173 is taken by something that isn't this day's page; stop it.
- **Still stuck?** Open `solution/fakeFetch.ts`, read it once, close it, then write your own version from memory.
