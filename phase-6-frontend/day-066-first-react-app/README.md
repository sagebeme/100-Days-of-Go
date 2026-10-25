# Day 66: First React App

## The brief

A first-week student at UoN stands at Railways and can't tell which mat goes to Ngong. Build them a page: a heading, then one card per route showing its number, where it goes, the SACCO, where it boards and the fare. Some routes have no number painted on, and some fares nobody has written down, so the cards must cope with both.

```
<FareBadge amountKsh={150} />       →  Ksh150
<FareBadge />                       →  Ask the conductor
<RouteCard route={route58} />       →  <article> <h2>58 to Buruburu</h2> Buruburu 58 Travellers Sacco · Boards at Ambassadeur · Ksh150 </article>
<RouteCard route={kiberaRoute} />   →  <article> <h2>To Kibera</h2> ... Ask the conductor </article>
<App routes={ROUTES} />             →  <h1>Matatu routes from the CBD</h1> and six cards
<App routes={[]} />                 →  <h1>...</h1> No routes yet.
```

The routes are real (checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026). Routes change, so confirm with the conductor.

## What you'll use

**Vite** serves the folder during development and builds it into static files for production. `index.html` loads `main.tsx`, which hands `<div id="root">` to React.

**A component** is a function whose name starts with a capital letter. It takes one object, its **props**, and returns **JSX**: HTML-like syntax that compiles to function calls. Curly braces drop back into TypeScript.

```tsx
function Greeting({ name, times }: { name: string; times?: number }) {
  return <p>Habari {name}{times ? ` x${times}` : ""}</p>;
}

<Greeting name="Wanjiku" />            // props are typed like function arguments
{routes.map((r) => <li key={r.number}>{r.goesTo}</li>)}  // a list is an array of JSX
{routes.length === 0 ? <p>None</p> : <ul>...</ul>}       // no if inside JSX: use ?: or &&
```

Compared with Go's `html/template`: here the template is code, so the compiler checks every prop name and type. React escapes text in `{}` for you, as `html/template` does, so a route called `<script>` shows as text.

`key` tells React which list item is which. Tomorrow explains it; today give each `<li>` a key that's unique in the list.

## Steps

1. Open `starter/App.tsx`. The three components return empty elements. Keep their names and props.
2. `FareBadge`: return a `<span>` holding `Ksh${amountKsh}`, or `Ask the conductor` when `amountKsh` is `undefined`. Test for `undefined`, not falsiness, so a Ksh0 fare would still show.
3. `RouteCard`: an `<article>` with an `<h2>` title (`58 to Buruburu`, or `To Kibera` without a number), a `<p>` with the operator, a `<p>` with `Boards at ...`, and `<FareBadge amountKsh={route.fareKsh} />`.
4. `App`: a `<main>` with the `<h1>`, then either `No routes yet.` or a `<ul>` with `routes.map(...)` giving an `<li key=...>` around each `RouteCard`. The 46K has a number, Kibera doesn't, so build the key from fields every route has.
5. Run the tests from `phase-6-frontend`, then open the page:

```bash
npm test -- day-066
npm run dev -- day-066-first-react-app/starter
```

Open http://localhost:5173. Edit a component and save: the page updates without a reload.

## The kata: Component Tree

Print a React-style component tree from child-parent pairs, the way DevTools shows it. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 66
```

## When you're stuck

- **`Unable to find an accessible element with the role "article"`**: the test looks for an `<article>` element. A `<div>` doesn't count.
- **The heading text is `58to Buruburu`**: JSX drops spaces at line breaks. Build the title as one template string.
- **`Warning: Each child in a list should have a unique "key" prop`**: put `key` on the outermost element inside `map`, the `<li>`, not on `RouteCard` inside it.
- **`'>' expected`** or other odd errors on JSX: JSX only works in `.tsx` files, not `.ts`.
- **`npm run dev` says port 5173 is in use**: another dev server is running. Stop it with Ctrl+C in its terminal.
- **Still stuck?** Open `solution/App.tsx`, read it once, close it, then write your own version from memory.
