# Day 74: Fast on a Cheap Phone

## The brief

Most of Twende's passengers have an entry-level Android phone that cost around Ksh8,000, a few hundred MB of bundles, and a 3G signal that comes and goes between Railways and Ngong. On your laptop the trip board feels instant. On their phone it takes seconds before anything works. Make it light: download less, draw less, redo less work, size images for the screen, and let the browser cache what doesn't change.

```
formatBytes(182_449)                                  →  "182.4 kB"
checkBudget([index.js 150 kB, SeatMap.js 32.4 kB], { js: 170_000, css: 30_000 })
                                                      →  ["JS is 182.4 kB, over the 170.0 kB budget"]
pickWidth(360, 2.75, [320, 640, 960])                 →  960     (a 360px slot on a 2.75x screen needs 990px)
srcSet("matatu", [960, 320, 640])                     →  "/img/matatu-320.webp 320w, /img/matatu-640.webp 640w, ..."

<TripBoard trips={600 trips} />                       →  "Showing 50 of 600 trips", 50 rows drawn
tick "Show seat map"                                  →  SeatMap-[hash].js downloads now, not on page load

GET /assets/index-DOzQj1Vn.js   →  Cache-Control: public, max-age=31536000, immutable
GET /                           →  Cache-Control: no-cache   (then 304 Not Modified on a re-check)
```

The departures are made up; the routes are real (checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026). Routes change, so confirm with the conductor.

## What you'll use

**Bundle size.** Every kilobyte of JavaScript must be downloaded, then parsed and run, and a cheap phone's CPU is several times slower than your laptop's. `npm run build` prints each file's size, raw and gzipped. Set a **budget** and check it, the way `checkBudget` does, so growth is a decision rather than an accident.

**Code splitting.** `React.lazy` with a dynamic `import()` tells Vite to put a component in its own file, downloaded the first time it renders. Wrap it in `<Suspense fallback={...}>` to show something meanwhile. Split things most people never open: the seat map, an admin screen, a chart library.

```tsx
const SeatMap = lazy(() => import("./SeatMap")); // SeatMap.tsx must have a default export
{showMap && <Suspense fallback={<p>Loading seat map...</p>}><SeatMap /></Suspense>}
```

**Less work per render.** React re-runs your component on every state change. `useMemo(() => search(trips, query), [search, trips, query])` keeps the last result until one of those changes. Draw only what fits on screen (50 rows, not 600); there are "virtual list" libraries for really long lists. Measure first with the React DevTools Profiler; `useMemo` everywhere just adds noise.

**Images.** Usually the heaviest thing on a page. `srcSet` lists the sizes you have and `sizes` says how wide the slot is, so the browser picks the smallest file that looks sharp. Use WebP or AVIF, `loading="lazy"` for images below the fold, and always `width` and `height` so the page doesn't jump when the image arrives.

**Caching headers from Go.** Vite puts a content hash in each asset's name, so a changed file gets a new name. That makes the assets safe to cache "forever" (`max-age=31536000, immutable`), while `index.html` gets `no-cache` so a new deploy is seen at once. `http.FileServer` already answers `If-Modified-Since` with an empty `304`. See `api/main.go`, and [http.md](../../reference/http.md) for caching in general.

## Why it matters at work

Slow pages lose people: on a metered connection, a page that downloads 2 MB before showing a trip costs the passenger money, and many will give up before it loads. Performance budgets in CI catch the one dependency that doubles the bundle, and a reviewer will ask why a date library is in the first download. Wrong cache headers cause the worst kind of incident: users stuck on an old version, or a CDN serving one user's page to another.

## Steps

1. In `starter/perf.ts`, write `formatBytes` (`toFixed(1)` for kB and MB), `checkBudget` (sum the `.js` and `.css` sizes with `filter` and `reduce`), `pickWidth` (sort a copy, then `find` the first wide enough, else the last) and `srcSet`. Run the tests: `perf.test.ts` should pass.
2. In `starter/App.tsx`, declare `const SeatMap = lazy(() => import("./SeatMap"))` at the top of the file, outside the component.
3. In `TripBoard`, keep `query` and `showMap` in state. Work out the results with `useMemo`, and draw only the first 50.
4. Render the `<img>` with `srcSet`, `sizes`, `loading="lazy"`, `width={640}` and `height={360}`, the search box (`type="search"`), the checkbox, the `Suspense`-wrapped `SeatMap` when ticked, `Showing N of M trips` and the list.
5. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-074
```

6. Build it and look at the output. You should see a separate `SeatMap-[hash].js`:

```bash
npm run build -- day-074-fast-on-a-cheap-phone/starter
```

7. Serve the build from Go, from the repo root, then open http://localhost:8080:

```bash
go run ./phase-6-frontend/day-074-fast-on-a-cheap-phone/api
```

8. In Chrome DevTools, open Network, make sure "Disable cache" is unticked, pick a throttling preset such as 3G, and reload twice. The second load takes the JS from the cache, and the HTML comes back 304. Tick "Show seat map" and watch its chunk arrive. Run Lighthouse with mobile settings for a score. The picture's files aren't in the repo, so it shows its alt text; that's expected.

To see the page during development instead: `npm run dev -- day-074-fast-on-a-cheap-phone/starter`.

## The kata: Initial Bundle

From a module graph with static and lazy imports, work out the first download, what can wait, and what's never used. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 74
```

## When you're stuck

- **`Element type is invalid. Received a promise that resolves to: undefined`**: `lazy` needs the module's **default** export. `SeatMap.tsx` has `export default function SeatMap`.
- **The page goes blank, or React complains that a component suspended**: the lazy component has no `<Suspense fallback={...}>` above it.
- **No `SeatMap` chunk in the build output**: something imports `./SeatMap` statically too, so it's already in the main bundle. Only `lazy(() => import(...))` may import it.
- **"toggling the seat map doesn't search again" fails**: the search runs in the component body. Wrap it in `useMemo` with `[search, trips, query]`.
- **`npm run build` warns `Some chunks are larger than 500 kB`**: that's Vite's own budget. Split with `lazy`, or find the heavy dependency with a bundle visualiser.
- **The Go server says `no such file or directory`**: build first; it serves `starter/dist`. Set `STATIC_DIR` to serve another folder.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
