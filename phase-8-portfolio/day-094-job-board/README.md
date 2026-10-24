# Day 94: Job Board

## The brief

Wairimu runs a WhatsApp group for 900 junior developers in Nairobi. Every day someone posts a job, and by evening it's buried under 300 messages. She wants Kazi Board (a made-up name): a site where companies post jobs and juniors search them, plus a JSON API so her friend can build a Telegram bot on top. It has to work on a cheap phone with JavaScript off, and a stranger's site must not be able to post jobs through a visitor's browser.

```
POST /api/jobs  {"title":"Go Developer","company":"Twende","location":"Nairobi","tags":["Go"]}
                                   →  201  Location: /api/jobs/1
POST /api/jobs  {"title":"Go"}     →  422  {"error":"invalid job","fields":{"title":"must be 3 to 80 characters",...}}
GET  /api/jobs?q=go+backend&page=1 →  200  {"jobs":[...],"page":1,"per_page":10,"total":3,"pages":1}
GET  /?q=go                        →  200  an HTML page: a search box, ten results, a Next link
POST /jobs  title=...&company=...  →  303  Location: /jobs/2
POST /jobs  Sec-Fetch-Site: cross-site
                                   →  403
```

**Acceptance criteria**

- Jobs are validated on the server, and every problem is reported at once, per field.
- Search splits the query into words and ranks jobs by where they match: title 3 points, tag 2, company 1. Ties go to the newest job.
- Results come ten to a page, in JSON and in HTML.
- HTML pages escape everything a user typed. Cross-site POSTs are refused.

## What you'll use

- JSON in and out, `201` with `Location`, `415`, from [Day 52: JSON API](../../phase-5-backend/day-052-json-api/).
- `http.MaxBytesReader`, `400` against `422`, from [Day 53: Validation](../../phase-5-backend/day-053-validation/).
- `http.CrossOriginProtection` and why `SameSite` isn't the whole story, from [Day 61: CORS and Security Headers](../../phase-5-backend/day-061-cors-and-security-headers/).
- XSS and CSRF from [Day 60: OWASP Top 10](../../phase-5-backend/day-060-owasp-top-10/) and [owasp-checklist.md](../../reference/owasp-checklist.md). Status codes from [http.md](../../reference/http.md).
- Custom sort orders, from [Day 31: Sorting Showdown](../../phase-3-algorithms/day-031-sorting-showdown/).
- New today: `html/template`. It looks like `text/template` but knows where in the page each value lands, and escapes it to match: `<` becomes `&lt;` in text, and a value in an `href` gets URL-escaped.

```go
var page = template.Must(template.New("job").Parse(`<h1>{{.Title}}</h1>`))
page.Execute(w, Job{Title: "<script>alert(1)</script>"}) // <h1>&lt;script&gt;alert(1)&lt;/script&gt;</h1>
```

## Steps

Requirements. The tests in `starter/` check every must-have; the doc comments give the exact rules.

**Must-haves**

1. `Normalize` and `Validate`: trim, lower-case and dedupe tags, then check every field and return a `ValidationError` with one message per bad field.
2. `Tokenize`: lower-case, then split on anything that isn't a letter or digit (`unicode.IsLetter`, `unicode.IsDigit`). `M-Pesa` is two words.
3. `Score` and `Search`: whole-word matches only, each distinct query word counted once. Sort by score, then by ID, both descending. Don't reorder the caller's slice.
4. `Paginate`: clamp a bad page or page size, and return `[]` (not `null` in JSON) past the end.
5. `Store`: an in-memory slice behind a `sync.RWMutex`. IDs start at 1, and a rejected job doesn't use one.
6. `NewServer`, the JSON routes: `DisallowUnknownFields`, a body limit of `MaxBodyBytes` (`413` when it's hit), `415` for anything but `application/json`, `422` with `fields` for a bad job, and the client can't choose the ID.
7. `NewServer`, the HTML routes: a search page with a form that uses `GET` (so results can be bookmarked) and Previous and Next links that keep `q`; a job page; a form page. A bad form comes back with `422`, the user's values filled in and the messages next to the fields. A good one gets `303 See Other`, so refreshing doesn't post twice.
8. Wrap the whole mux: `return http.NewCrossOriginProtection().Handler(mux)`.
9. Run the tests:

```bash
go test ./phase-8-portfolio/day-094-job-board/starter
```

**Stretch goals**

- A `cmd/kazi` main that seeds ten jobs and serves on `PORT` (Day 21's config rules).
- Search suggestions for the search box, from the kata.
- Prefix matching (`dev` finds `developer`) worth fewer points than a whole word.
- Store jobs in Postgres (Day 55) behind an interface, and search with `tsvector`. Keep the in-memory store for tests.
- Security headers from Day 61, including a `Content-Security-Policy` that allows no inline scripts.

## The kata: Search Suggestions

Given job titles and prefixes, print the first three titles in alphabetical order for each prefix. It's LeetCode's "search suggestions system", a common interview question, solved with sort and binary search (or a trie). Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 94
```

## When you're stuck

- **`pattern "GET /jobs/new" conflicts with pattern "GET /jobs/{id}"`**: it doesn't, in Go 1.22 and later: the more specific pattern wins. If you see a conflict, check that both use the same method.
- **`GET /` answers every unknown path with the search page**: `/` matches everything. Use `GET /{$}` for exactly `/`.
- **The API returns `"jobs":null`**: a nil slice encodes as `null`. Start with `[]Job{}`.
- **`superfluous response.WriteHeader call`**: you wrote a header after the body started. Execute the template into a `strings.Builder` or `bytes.Buffer` first, then write the status and the bytes. That also turns a template error into a clean `500`.
- **The form test gets `403`**: you're sending `Sec-Fetch-Site: cross-site` or an `Origin` that isn't the host. Browsers set those headers; plain test requests don't, and `CrossOriginProtection` lets those through.
- **A `{{.Query}}` with `&` breaks the Next link**: build the link inside the template's `href`, where `html/template` escapes for URLs. Don't build HTML strings in Go and mark them `template.HTML`.
- **Scores look right but the order is wrong**: compare `b` to `a` for descending, and break ties on ID inside the same compare function.
- **Still stuck?** Open `solution/jobs.go` and `solution/server.go`, read them once, close them, then write your own version from memory.

## Make it portfolio-ready

- **README**: a screenshot of the search page on a phone, the API table from the `NewServer` doc comment with one `curl` example per route, and the ranking rule in one sentence.
- **Demo and deploy**: add the `cmd/kazi` main and a Dockerfile (Day 63), deploy it to a free tier (Render, Fly.io or Koyeb), and put the link at the top of the README. Seed it with fake jobs, never real people's details.
- **Tests and CI**: `go test -race ./...` and `go vet` in GitHub Actions, with the badge in the README. Add one end-to-end test with `httptest.NewServer` that posts a form and follows the redirect.
- **What you'd do next**: write down the limits. An in-memory store loses everything on restart, and ranking by scanning every job is O(n) per search. Say what you'd change at 100 000 jobs: an inverted index (word to job IDs), or Postgres full-text search.
