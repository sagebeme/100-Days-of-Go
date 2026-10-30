# Day 75: Deploying the Front End

## The brief

The booking page is ready to leave your laptop. A React build is just files (`index.html`, a few hashed `.js` and `.css` files, images), so any static host can serve it, close to passengers and cheaply. What changes between your laptop, staging and production is **configuration**: which API to call, which features are on. Today you write the one module that reads that config, checks it, and refuses to start with a bad one. You also learn the rule that catches people out on their first deploy: every `VITE_` variable ends up in the JavaScript, readable by anyone.

```
parseConfig({})                         →  development, API http://localhost:8080, WS ws://localhost:8080/ws/seats
parseConfig({ VITE_APP_ENV: "production", VITE_API_URL: "https://api.twende.example/" })
                                        →  production, API https://api.twende.example, WS wss://api.twende.example/ws/seats
parseConfig({ VITE_APP_ENV: "staging", VITE_API_URL: "http://staging.twende.example" })
                                        →  throws: VITE_API_URL must use https in staging
findLeaks({ VITE_MPESA_CONSUMER_SECRET: "...", JWT_SECRET: "..." })
                                        →  ["VITE_MPESA_CONSUMER_SECRET"]

GET /config.json        →  {"apiUrl":"https://other.twende.example"}   (from the server's env, at runtime)
GET /bookings/K7Q2M9XA  →  index.html, so React can draw that page
GET /assets/nope.js     →  404
```

## What you'll use

**Build-time config.** Vite replaces `import.meta.env.VITE_API_URL` with the actual string when it builds. Values come from your shell or from `.env` files in the folder being built (`.env`, `.env.production`, or `.env.staging` with `vite build --mode staging`). Only names starting with `VITE_` are exposed, and they're exposed to everyone: build with a made-up `VITE_SECRET_THING=leaked123` and you'll find `leaked123` in `dist/assets/*.js`. Secrets (M-Pesa keys, JWT secrets) live only on the Go server. See [env-and-config.md](../../reference/env-and-config.md).

**Runtime config.** Build-time values mean one build per environment. To promote the *same* tested build from staging to production, the page can fetch `/config.json` at startup, and the server fills it from its own environment. `main.tsx` does this: build-time config first, then `mergeRuntime` with whatever `/config.json` says.

**Fail fast, all at once.** Like Day 21's Go config, check everything at startup and report every problem together, so one failed deploy shows all of them. A page silently calling `http://localhost:8080` from production is far worse than a page that refuses to start.

**Static hosting.** Netlify, Vercel, Cloudflare Pages, GitHub Pages, or S3 behind a CDN all do the same three things, which `api/main.go` does in a few lines of Go:

1. serve files, with long caching for hashed `assets/` and `no-cache` for `index.html` (Day 74)
2. **SPA fallback**: a path with no file, like `/bookings/K7Q2`, gets `index.html`, so a refresh or a shared link doesn't 404. On Netlify that's a `_redirects` line, `/* /index.html 200`.
3. HTTPS, which the hosts do for you. Behind your own Go server, that's a reverse proxy or `autocert`.

## Why it matters at work

Secrets in front-end bundles are one of the most common real leaks: an API key in a `VITE_` or `REACT_APP_` variable gets scraped from the JavaScript within days of going live. Reviewers will ask where each config value comes from, whether it's public, and what happens if it's missing. Config checked at startup turns a 2 a.m. "bookings are failing" incident into a deploy that never went out.

## Steps

1. In `starter/config.ts`, write `parseConfig`. Read each variable, treat `""` as unset, collect problems in an array, and throw one `ConfigError(problems)` at the end. `new URL(s)` throws for a string that isn't a URL, and `.protocol` gives `"https:"`. Build the WebSocket URL with `apiUrl.replace(/^http/, "ws") + "/ws/seats"`.
2. Write `findLeaks`: keys that start with `VITE_` and match `/SECRET|PASSWORD|PRIVATE|TOKEN/i`, sorted.
3. Write `mergeRuntime`: reject anything that isn't a plain object, copy `apiUrl`, `wsUrl` and `seatMap` when they have the right type, then run the same URL checks as `parseConfig`. Put the shared checks in one helper so the rules can't drift apart.
4. In `starter/App.tsx`, `EnvBadge` shows the mode in capitals with `role="note"`, and nothing in production. `App` shows the badge, the heading, the API URL and the seat map flag.
5. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-075
```

6. Build for staging and find your config, and the fake secret, in the output:

```bash
VITE_APP_ENV=staging VITE_API_URL=https://staging.twende.example VITE_SECRET_THING=leaked123 \
  npm run build -- day-075-deploying-the-front-end/starter
grep -l leaked123 day-075-deploying-the-front-end/starter/dist/assets/*.js
```

7. Serve that build like a static host, overriding the API at runtime, from the repo root. Open http://localhost:8080 and then http://localhost:8080/bookings/anything:

```bash
API_URL=https://other.twende.example go run ./phase-6-frontend/day-075-deploying-the-front-end/api
```

To see the page during development instead: `npm run dev -- day-075-deploying-the-front-end/starter`. Try `VITE_APP_ENV=production npm run dev -- ...` to watch it refuse to start.

## The kata: Deploy Diff

Compare what's live with a new build and work out what to upload, what to delete, and which old assets to keep for passengers still on the previous version. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 75
```

## When you're stuck

- **`import.meta.env.VITE_API_URL` is `undefined` in the build**: the variable was set after the build, or without the `VITE_` prefix, or in a `.env` file outside the folder being built. Build-time means the value must exist when `npm run build` runs.
- **A refresh on `/bookings/K7Q2` gives the host's 404 page**: the host has no SPA fallback. Add the rewrite to `index.html`.
- **The page loads a blank screen after a deploy**: `index.html` was cached and points at assets that were deleted. Serve HTML with `no-cache`, and keep old assets around (that's the kata).
- **`TypeError: Invalid URL`** escaping from your code: `new URL` throws. Wrap it in `try`/`catch` and turn it into a problem message.
- **"every problem at once" lists only one**: you threw on the first problem. Push to the array and throw once at the end.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
