# Phase 6 recap: Front end with TypeScript and React

Days 65 to 76. You put a face on your Go API: a TypeScript and React app that copes with slow networks, validates like the server, keeps tokens safe, updates live, works with a keyboard and a screen reader, loads fast on a cheap phone, and deploys as static files.

## What the phase covered

| Days | Topic |
|---|---|
| 65 | TypeScript for Go developers: types, interfaces, unions, `unknown`, `async` and `Promise.all` |
| 66 to 67 | Vite, components and props; `useState`, events, controlled inputs, lists and keys |
| 68 | `fetch`, loading and error states, `useEffect` with cleanup and `AbortController`, CORS in practice |
| 69 to 70 | forms validated on both sides, 422 field errors; tokens in memory, a refresh cookie, one shared refresh |
| 71 | WebSockets from Go to React, reconnecting with backoff, versions |
| 72 to 73 | accessibility: native elements, names, focus, contrast; tests with Vitest, Testing Library, fakes and Playwright |
| 74 to 75 | bundle size, code splitting, images and cache headers; build-time and runtime config, static hosting |
| 76 | the capstone: the booking app, end to end against the Go API |

## Key ideas

1. **Types vanish at run time.** Data from the network is `unknown` until you've checked it; a discriminated union makes every state explicit and lets the compiler check you handled each one.
2. **State is immutable.** React re-renders when the value changes by `===`, so make new arrays and objects; use the function form of a setter when the new value depends on the old one.
3. **Keys identify items, not positions.** Use a stable ID, or React gives one item's state to another when the list changes.
4. **Every fetch has three states and a cleanup.** Loading, error with a retry, done; abort the request the effect started, and ignore the abort error.
5. **The browser checks for speed, the server for safety.** Same rules, same messages, and always show the server's field errors.
6. **Tokens: the long one where scripts can't reach, the short one in memory.** `HttpOnly` cookie for refresh, a variable for access, one refresh in flight shared by every caller.
7. **Accessible by default means native first.** Buttons, labels, fieldsets and radios give you keyboard use and names for free; move focus when the page changes under the user.
8. **Every `VITE_` value is public.** Secrets live only on the server; config is checked at startup, and a page that refuses to start beats one that quietly calls localhost.

## Self-test

1. `fetch` gets a 500. Does the promise reject? What must your code do?
2. Two `setCount(count + 1)` calls in one click: what's the count after one click, starting from 0? And with `setCount((c) => c + 1)` twice?
3. Why does `RequireAuth` not protect the bookings, and what does?
4. A list uses `key={index}`. The first item is removed. What can go wrong in the second item?
5. Why is `<div onClick>` worse than `<button>`? Name two things the `div` doesn't do.
6. You find `VITE_PAYMENT_SECRET` in `.env.production`. Who can read it?

<details>
<summary>Answers</summary>

1. No: `fetch` only rejects when there's no response at all. Check `res.ok` and throw, so the error state shows.
2. `1`, because both calls use the `count` from that render. `2` with the function form.
3. Anyone can change the JavaScript in their own browser. The API's 401 on the protected routes is the real protection; `RequireAuth` is only for convenience.
4. React reuses the first item's element for what was the second item, so any state inside it (an uncontrolled input, an open panel) moves to the wrong item.
5. It isn't focusable with Tab, doesn't activate with Enter and Space, isn't announced as a button, and doesn't submit a form.
6. Everyone: Vite writes it into the JavaScript bundle that every visitor downloads.

</details>

## Build it without steps: Stage Lost and Found

Matatu conductors find phones, IDs and umbrellas every day. Build a lost-and-found board for one SACCO: conductors post found items, passengers search and claim them. Write a small Go API for it (Phase 5 skills), then the front end in `phase-6-frontend` with today's tools.

Done when:

- The list page searches as you type, shows loading, error (with Try again) and empty states, and never shows the results of an older search after a newer one.
- A conductor logs in to post an item through a form that validates name, stage, date and description with the same rules and messages as the API, and shows the API's 422 errors under the right field.
- The access token is kept in memory only, the refresh token in an `HttpOnly` cookie, and three requests that hit an expired token share one refresh.
- New items appear on every open page within a second, over a WebSocket that reconnects with backoff after the connection drops.
- Everything works with only a keyboard, every control has a name, focus moves to each new page's heading, and the colours pass WCAG AA (checked in a test).
- Component tests use a fake `fetch` and role queries; at least one test fails if you break the search on purpose.
- The photo view is a separate chunk loaded only when opened, `npm run build` stays under a budget you set, and the Go server sends long-lived cache headers for hashed assets and `no-cache` for `index.html`.
- Config is read and checked at startup, no secret is in any `VITE_` variable, and a refresh on a deep link doesn't 404.
