# Kata: Deploy Diff

Twende's front end is a folder of files on a static host. Uploading the whole folder on every deploy wastes time and bundles, so the deploy script compares the files already live (the old manifest) with the new build (the new manifest) and only sends what changed.

There's one catch. Files under `assets/` have a content hash in their name, and a passenger who opened the page five minutes ago may still need the old `index-3f9a2c.js` or a lazy chunk that went with it. So old assets are never deleted on deploy; they're kept (and cleaned up much later).

For each path:

- in the new build and either not live yet or with a different hash: `upload PATH`
- live but not in the new build, outside `assets/`: `delete PATH`
- live but not in the new build, under `assets/`: `keep PATH`
- the same path and hash in both: unchanged

**Input**

- A line with `n`, then `n` lines `path hash`: the old manifest (0 ≤ n ≤ 5 000)
- A line with `m`, then `m` lines `path hash`: the new manifest (0 ≤ m ≤ 5 000)

Paths have no spaces. Each path appears at most once per manifest.

**Output**

All `upload` lines, then all `delete` lines, then all `keep` lines, each group sorted by path (byte order). Finish with `unchanged K`.

**Example**

```
Input                               Output
5                                   upload assets/index-7c1e4b.js
index.html aaaa1111                 upload img/matatu-640.webp
assets/index-3f9a2c.js 3f9a2c77     upload index.html
assets/SeatMap-b81d0e.js b81d0e12   delete img/old-banner.webp
robots.txt cccc3333                 keep assets/index-3f9a2c.js
img/old-banner.webp dddd4444        unchanged 2
5
index.html eeee5555
assets/index-7c1e4b.js 7c1e4b90
assets/SeatMap-b81d0e.js b81d0e12
robots.txt cccc3333
img/matatu-640.webp ffff6666
```

`index.html` changed, so it's uploaded. The seat map chunk and `robots.txt` didn't change. The old banner is deleted, but the old `index-3f9a2c.js` is kept for passengers still on the previous version.

**Limits:** 2 seconds per test case.

**Go tip:** read the old manifest into a `map[string]string`, then walk the new one. `strings.HasPrefix(path, "assets/")` decides delete or keep. Collect each group in its own slice and `sort.Strings` it.

Judge it from the repo root with `go run ./tools/kata 75`.
