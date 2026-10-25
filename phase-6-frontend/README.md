# Phase 6: Front end with TypeScript and React

Days 65 to 76 are TypeScript and React. Each day still ends with a Go kata.

Install the tools once (Node 24):

```bash
cd phase-6-frontend && npm ci
```

`npm ci` installs the exact versions in `package-lock.json` into `node_modules/` (gitignored). Then, from `phase-6-frontend`:

| Command | What it does |
|---|---|
| `npm test -- day-066` | run one day's tests (Vitest) |
| `npm run typecheck` | check every file with the TypeScript compiler |
| `npm run dev -- day-066-first-react-app/starter` | serve that folder in the browser at http://localhost:5173 |

Days that talk to Go have an `api/` folder. Start it from the repo root in a second terminal, for example `go run ./phase-6-frontend/day-068-talking-to-go/api`. It listens on http://localhost:8080 and allows the origin in `ALLOWED_ORIGIN` (default `http://localhost:5173`).

The tests never start a server. They swap `fetch` for a fake, so they are fast and work offline.

All files in a day: `starter/` (yours to edit, holds the tests), `solution/` (one way to do it), `kata/` (the Go puzzle), and sometimes `api/` (the Go server).
