# Day 67: State and Events

## The brief

At Railways stage, the conductor of a 14-seater keeps two things in his head: how many seats are taken, and who is waiting and where they're going. Put both on his phone. A seat counter with Board and Alight buttons that knows when the mat is full, and a queue where he adds names, notes each passenger's stop, and taps Board when they get in.

```
<SeatCounter capacity={14} />   click Board ×3, Alight ×1   →  "2 of 14 seats taken"
<SeatCounter capacity={2} />    click Board ×2              →  "2 of 2 seats taken", Board disabled, "Imejaa! The mat is full."

<StageQueue />   type "Amina", Add · type "Baraka", Add   →  In the queue: 2   [Amina] [Baraka]
                 type "Kibera" in Amina's stop, Board Amina →  In the queue: 1   [Baraka], with an empty stop box

addPassenger([], "  Chebet ", 1)                          →  [{ id: 1, name: "Chebet" }]
removePassenger([{id:1,"Kamau"}, {id:2,"Kamau"}], 2)      →  [{ id: 1, name: "Kamau" }]
```

## What you'll use

**`useState`** gives a component a value that survives re-renders, and a setter. Calling the setter makes React run the component again with the new value. A plain `let` would reset to its first value on every render.

```tsx
const [taken, setTaken] = useState(0);
<button onClick={() => setTaken((t) => t + 1)}>Board</button>  // the function form uses the latest value
```

**Never change state in place.** React compares the old and new value with `===`. `queue.push(p)` changes the same array, so React thinks nothing happened. Make a new one: `[...queue, p]` to add, `queue.filter(...)` to remove. That's why `queue.ts` holds pure functions you can test on their own.

**Events.** `onClick`, `onChange` and `onSubmit` take a function. A **controlled input** shows state and updates it: `value={name} onChange={(e) => setName(e.target.value)}`. Put the input in a `<form>` with `onSubmit`, and Enter works too; call `e.preventDefault()` so the browser doesn't reload the page.

**Keys.** When a list changes, React matches old and new items by `key`. With `key={index}`, boarding Amina makes Baraka's `<li>` "become" item 0, and he inherits what was typed in Amina's stop box. With `key={p.id}`, each `<li>` stays with its passenger. Use a stable, unique id; that's also why two Kamaus need different ids.

## Steps

1. In `starter/queue.ts`, write `addPassenger` (trim; a blank name returns the list unchanged; otherwise `[...list, { id, name }]`) and `removePassenger` (`filter` by id). Run the tests: `queue.test.ts` should pass.
2. In `starter/App.tsx`, `SeatCounter`: one `useState(0)`. Show `{taken} of {capacity} seats taken`, two buttons with `disabled={...}`, and the full message with `{full && <p role="status">...</p>}`.
3. `StageQueue`: state for the queue, the name in the box, and the next id. In the form's `onSubmit`, prevent the default, add the passenger, bump the id, clear the box.
4. Render `<ul aria-label="Queue">` with `queue.map(...)`. Each `<li key={p.id}>` holds `<span className="name">`, a `<label>Stop for {p.name} ({p.id}) <input /></label>` and a `Board {p.name}` button that removes them. The stop box is uncontrolled on purpose: it's what shows the key bug.
5. Try `key={i}` (with `map((p, i) => ...)`) once and watch the last test fail. Then put `p.id` back.
6. Run the tests from `phase-6-frontend`, then open the page:

```bash
npm test -- day-067
npm run dev -- day-067-state-and-events/starter
```

## The kata: Undo Button

Replay add, remove and undo actions on a queue. An undone remove puts the passenger back where they stood. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 67
```

## Check yourself

1. Why can't `SeatCounter` keep the count in a plain `let taken = 0`?
2. Predict what the counter shows after one click:

   ```tsx
   const [taken, setTaken] = useState(0);
   <button onClick={() => { setTaken(taken + 1); setTaken(taken + 1); }}>Board two</button>
   ```

   And what if both calls were `setTaken((t) => t + 1)`?
3. Spot the bug. Clicking Add does nothing on screen, though `console.log(queue)` shows the new passenger:

   ```tsx
   function add(name: string) {
     queue.push({ id: nextId, name });
     setQueue(queue);
   }
   ```

4. The queue renders with `key={i}`. Amina's stop box says "Kibera", then the conductor boards Amina. What does Baraka's stop box show, and why does `key={p.id}` fix it?

<details>
<summary>Answers</summary>

1. React calls the component function again on every render, so a `let` starts at 0 each time, and changing it doesn't tell React to re-render. `useState` keeps the value between renders, and its setter triggers the next render.
2. `1`. Both calls use the `taken` from this render (0), so both set it to 1. With the function form each update gets the latest value, so it shows `2`.
3. `push` changes the same array, so the old and new state are `===` and React skips the re-render. Make a new array: `setQueue([...queue, { id: nextId, name }])`.
4. "Kibera". With index keys, after Amina leaves, Baraka is item 0, so React reuses item 0's `<li>` (and its uncontrolled input) for him. With `key={p.id}`, React removes Amina's `<li>` and keeps Baraka's, so his box stays empty.

</details>

## When you're stuck

- **The count never changes, but no error**: you changed state in place (`push`, `splice`, `taken++`). Make a new value and pass it to the setter.
- **`Too many re-renders. React limits the number of renders`**: you wrote `onClick={setTaken(taken + 1)}`, which calls it during render. Pass a function: `onClick={() => setTaken(taken + 1)}`.
- **The page reloads when you press Add**: the form submitted for real. Call `e.preventDefault()` first in `onSubmit`.
- **"Baraka's stop after Amina boards" is `Kibera`**: your key is the index. Use `p.id`.
- **`Unable to find a label with the text of: Stop for Amina (1)`**: the label text must be exactly `Stop for NAME (ID)`, with the input inside the `<label>`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
