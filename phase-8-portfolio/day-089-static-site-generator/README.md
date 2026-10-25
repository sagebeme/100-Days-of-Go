# Day 89: Static Site Generator

## The brief

Kamau is a junior developer in Nakuru who wants a blog for his job hunt. WordPress hosting costs more than he wants to pay, and he'd rather show off Go anyway. The plan: write posts in Markdown, run one command, and push a folder of plain HTML to free static hosting. His friend Akinyi will write guest posts, so he wants it safe by default: whatever is in a post, it mustn't be able to run a script on his site unless he says so.

```
posts/matatu-fares.md:
  ---
  title: What a matatu ride costs
  date: 2026-10-24
  tags: nairobi
  ---
  | Route | To | Fare |
  ...
  <script>alert("hi")</script>

Build(site, Options{SiteTitle: "Kamau's Blog"})  →  index.html
                                                    matatu-fares/index.html
                                                    first-week-in-go/index.html
                                                    static/style.css
matatu-fares/index.html                         →  <h1>What a matatu ride costs</h1> ... <table> ... <!-- raw HTML omitted -->
posts/broken.md (no front matter)               →  error: broken.md: a post must start with --- front matter
two posts with slug "older"                     →  error: two posts have the same slug: ...
```

Acceptance criteria:

- Front matter gives the title, date, tags, draft flag and an optional slug. Unknown keys are errors, so typos don't vanish silently.
- Markdown goes through goldmark with tables. Raw HTML and `javascript:` links are left out unless the site owner turns on `AllowHTML`.
- Layouts are `html/template`, so titles and descriptions are escaped automatically.
- The index lists posts newest first. Drafts are left out unless asked for.
- A slug can never write outside the output folder.

`example/` has a small site to build: a layout, three posts (one a draft) and a stylesheet.

## What you'll use

- Splitting strings with `strings.Cut` and `strings.CutPrefix`: [Day 8](../../phase-1-terminal/day-008-hashtag-maker/).
- Sorting with a comparator: [Day 31](../../phase-3-algorithms/day-031-sorting-showdown/).
- Files and folders: [Day 23](../../phase-2-modelling/day-023-save-load/). Flags for the CLI: [Day 22](../../phase-2-modelling/day-022-cli-flags/).
- XSS and output encoding from [Day 60](../../phase-5-backend/day-060-owasp-top-10/) and [reference/owasp-checklist.md](../../reference/owasp-checklist.md).
- New today: `github.com/yuin/goldmark`, `html/template` layouts with `{{define}}`, and `io/fs`. `Build` takes an `fs.FS`, so the real program passes `os.DirFS(dir)` and the tests pass an in-memory `fstest.MapFS`.

```go
md := goldmark.New(goldmark.WithExtensions(extension.GFM))
md.Convert(src, &buf) // raw HTML becomes <!-- raw HTML omitted -->
template.HTML(buf.String()) // tells html/template: already safe, don't escape again
```

## Why it matters at work

`template.HTML` is the one place where `html/template` stops protecting you, so code review treats every conversion to it as a security decision. Here it's safe only because goldmark has already dropped raw HTML. Turn on `WithUnsafe` and a guest post can carry a `<script>` that runs for every visitor. The same reasoning applies to any app that renders user Markdown: comments, READMEs, support tickets. The other habit to take away is to depend on `fs.FS`, not on the disk, so a whole build is tested in memory in milliseconds.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `ParseFrontMatter` follows its doc comment, including errors for no front matter, no closing `---`, a line without a colon and unknown keys.
2. `Slugify` and `ValidSlug`. A slug from the front matter is checked too: `../../etc` is refused.
3. `RenderMarkdown` with the GFM extension, safe unless `allowHTML`.
4. `ParsePage` fills a `Page`: dates as `YYYY-MM-DD`, `draft` with `strconv.ParseBool`, tags split, trimmed and lower-cased. Every error names the file.
5. `Build` parses `layout.html` and checks it defines `page` and `index`. It reads `posts/*.md`, skips drafts unless asked, refuses duplicate slugs, sorts, renders each page and the index, and copies `static/` as it is. A site without `static/` still builds.
6. `Write` refuses any path that isn't local (`filepath.IsLocal`) and makes folders as it goes.
7. `cmd/ssg` builds a folder, and with `-serve :8080` serves the result for a preview.

```bash
go test ./phase-8-portfolio/day-089-static-site-generator/starter
go run ./phase-8-portfolio/day-089-static-site-generator/starter/cmd/ssg \
    -src phase-8-portfolio/day-089-static-site-generator/example -out /tmp/kamau-blog -serve :8080
```

Stretch goals:

- Tag pages at `tags/<tag>/index.html`, and an RSS feed at `feed.xml` with `encoding/xml`.
- A `-watch` flag that rebuilds when a file changes (poll modification times every second).
- Syntax highlighting for code blocks (it needs a goldmark extension, so it's a new dependency: ask first at work).
- Reading time on each post, counted from the words in the Markdown.

## The kata: Tag Balance

Check that the tags in a line of HTML are balanced, and say where the first problem is. It's Day 33's bracket checker with tag names. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 89
```

## When you're stuck

- **The page shows `&lt;p&gt;Hello&lt;/p&gt;` as text**: you put the rendered body in a `string` field. `html/template` escapes strings. Store it as `template.HTML`.
- **`<script>` appears in the output**: you used `html.WithUnsafe()` all the time. Only add it when `allowHTML` is true.
- **`html/template: "page" is undefined`**: the layout has no `{{define "page"}}`. Check for it after parsing and return a clear error.
- **The index order changes between runs**: posts on the same date are tied. Break the tie with the slug.
- **`posts/sub/deep.md` gets built**: `fs.Glob(src, "posts/*.md")` doesn't go into sub-folders, which is what the tests want. If you used `WalkDir`, skip folders below `posts`.
- **The CRLF test fails with a title of `"Habari\r"`**: replace `"\r\n"` with `"\n"` before you split lines.
- **`open static: file does not exist`**: a site without `static/` is fine. In the `WalkDir` callback, return `fs.SkipDir` for that error on the root.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Use it to build your real blog. Put the generator in one repo and the posts in another, or both in one with a `content/` folder.
- Deploy the output to GitHub Pages, Netlify or Cloudflare Pages with a GitHub Actions workflow, and link the live site at the top of the README.
- Screenshot a post and the index for the README, and show the `example/` folder next to the HTML it produces.
- Write "What I'd do next": tags, RSS, watch mode, a sitemap.
- Add a CI badge for `go test` and a second job that builds `example/` and fails on any error.
