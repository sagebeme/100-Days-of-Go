# Day 72: Accessible by Default

## The brief

Mzee Otieno books his trips with a screen reader. Wanjiru's thumb is in a cast, so she uses a keyboard on her laptop. Half of Nairobi reads their phone in bright sun at a stage. Rebuild the booking form so it works for all of them: every field named, every seat a real radio button, errors that announce themselves and take you to the problem, a confirm dialog that holds focus and gives it back, and colours that pass WCAG contrast.

```
contrastRatio("#767676", "#ffffff")  →  4.54   meetsAA(4.54, false)  →  true
contrastRatio("#777777", "#ffffff")  →  4.48   meetsAA(4.48, false)  →  false

Tab to the seats, Space            →  "Seat 1, radio button, not checked, 1 of 6" → checked
Book with nothing filled in        →  focus jumps to "There is a problem": [Enter your name] [Choose a seat]
click "Enter your name"            →  focus moves to the Name box
Book with a name and seat 3        →  dialog "Confirm your booking", focus on Confirm
Tab, Tab                           →  Cancel, then Confirm again (focus stays in the dialog)
Escape                             →  dialog closes, focus back on Book
```

## What you'll use

**Native elements first.** A `<button>` is focusable, works with Enter and Space, and announces itself as a button. A `<div onClick>` does none of that. Radios in a `<fieldset>` with a `<legend>` give you arrow keys, Space and "radio button, 3 of 14" for free. Reach for ARIA (`role`, `aria-*`) only when HTML has nothing that fits, like a dialog's `aria-modal`.

**Names.** Every control needs an accessible name: a `<label htmlFor>` for inputs, the text inside a button, a `<legend>` for a group, `aria-labelledby` pointing at a heading for a dialog. Testing Library's `getByRole("radio", { name: "Seat 3" })` finds elements the way a screen reader does, so if your test can find it by role and name, so can Mzee Otieno.

**Focus management.** When the page changes under the user, move focus to the change, or they won't know it happened. Error summary: focus it (it needs `tabIndex={-1}` to accept focus) and give it `role="alert"`. Dialog: focus inside on open, keep Tab inside, close on Escape, return focus to the button that opened it. In React, move focus in a `useEffect` after the element exists, through a `useRef`.

```tsx
const summaryRef = useRef<HTMLDivElement>(null);
useEffect(() => { if (hasErrors) summaryRef.current?.focus(); }, [hasErrors]);
<div role="alert" tabIndex={-1} ref={summaryRef}>...</div>
```

**Contrast.** WCAG 2 AA needs a ratio of 4.5:1 for normal text and 3:1 for large text (24px, or about 18.66px bold). `contrast.ts` implements the formula, and the `THEME` tests in `contrast.test.ts` catch a designer's lovely pale grey before it ships.

**axe.** [axe-core](https://github.com/dequelabs/axe-core) is the engine behind most automated checkers (the axe DevTools browser extension, Lighthouse's accessibility audit). It finds missing labels, bad contrast, duplicate ids and broken ARIA in seconds. It can't tell whether focus goes somewhere sensible or whether a label makes sense, so it complements keyboard and screen reader testing; it doesn't replace them. Run Lighthouse in Chrome DevTools on your page today.

## Why it matters at work

Accessibility bugs are bugs: a form a keyboard user can't submit is broken for them. Many organisations must meet WCAG 2.1 AA by law or contract, and reviewers will ask why a clickable `<div>` isn't a `<button>`. Writing tests with role queries makes the accessible way the easy way, because the inaccessible markup is the one the test can't find.

## Steps

1. In `starter/contrast.ts`, write `relativeLuminance` (parse the three hex pairs with `parseInt(h, 16)`, linearise, weight), `contrastRatio` and `meetsAA`. Run the tests: `contrast.test.ts` checks the formula and then the theme's colours.
2. In `starter/App.tsx`, `SeatPicker`: a `<fieldset>` and `<legend>Choose a seat</legend>`, then for each seat a `<label>` wrapping `<input type="radio" name="seat">` and the text `Seat N` (plus ` (taken)` and `disabled` when taken).
3. `BookingForm`: labelled `Name` and `Phone` inputs, the `SeatPicker`, and a `Book` button. Add `noValidate` to the form.
4. On a bad submit, render the error summary and focus it in an effect. Each link's `onClick` prevents the default jump and focuses the field. Give each invalid field `aria-invalid` and an `aria-describedby` pointing at its message.
5. On a good submit, render the dialog. Focus Confirm in an effect. In the dialog's `onKeyDown`, close on `Escape`, and wrap `Tab` from the last button to the first and `Shift+Tab` the other way. Closing puts focus back on Book through a ref.
6. Confirm writes `Seat N is yours, NAME.` into a `<p role="status">` that is always on the page; screen readers announce changes to its text.
7. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-072
```

8. Open it with `npm run dev -- day-072-accessible-by-default/starter`. Put the mouse away and book a seat with only Tab, Shift+Tab, arrows, Space, Enter and Escape. Then run Lighthouse (Accessibility) in DevTools. If you can, try a screen reader: NVDA on Windows, VoiceOver on a Mac or iPhone, TalkBack on Android.

## The kata: Contrast Check

Compute WCAG contrast ratios for a designer's colour pairs and give each a verdict. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 72
```

## When you're stuck

- **`Unable to find role="group" and name "Choose a seat"`**: a `<fieldset>` is a group, and its name comes from a `<legend>` that is its first child.
- **`Unable to find an accessible element with the role "radio" and name "Seat 3"`**: the text must be inside the `<label>` that wraps the input, and nothing else can be in there.
- **"focus moves to the summary" fails**: a `<div>` can't take focus without `tabIndex={-1}`, and calling `focus()` during render does nothing because the element doesn't exist yet. Use a ref and an effect.
- **Focus lands on `<body>` after Escape**: the dialog's buttons are gone, so the browser dropped focus. Focus Book yourself in `close()`.
- **"Shift+Tab from the first wraps to the last" fails**: check `e.shiftKey`, and call `e.preventDefault()` before moving focus, or the browser moves it again.
- **`3.00` for a ratio that fails**: the theme test compares the full number. 2.996 rounds to 3.00 but is below 3.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
