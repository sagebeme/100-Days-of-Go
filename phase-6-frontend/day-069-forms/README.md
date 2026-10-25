# Day 69: Forms

## The brief

Passengers can find a route; now they book seats. The form asks for a name, a phone number for the SMS, a route and how many seats. People type phone numbers every way possible, forget fields and ask for 20 seats in a 14-seater. Catch what you can in the browser, instantly, with exactly the same messages the Go API uses. Some things only the server knows, like the 111 having three seats left, so show its 422 errors under the right field too.

```
normalisePhone("0712 345 678")    →  "254712345678"
normalisePhone("+254110123456")   →  "254110123456"
normalisePhone("0812345678")      →  null
validateBooking({ name: " ", phone: "12", route: "", seats: 0 })
  →  { name: "Enter your name", phone: "Enter a Kenyan mobile number like 0712 345 678",
       route: "Choose a route", seats: "Book between 1 and 14 seats" }

POST /api/bookings {"name":"Amina","phone":"0712345678","route":"58","seats":2}
  →  201 {"reference":"K7Q2M9XA","phone":"254712345678","route":"58","seats":2}
POST /api/bookings {"name":"Kip","phone":"0712345678","route":"111","seats":4}
  →  422 {"errors":{"seats":"Only 3 seats left on 111"}}
```

Seats left are made up for the exercise.

## What you'll use

**Controlled inputs.** The input shows state, and every keystroke updates state, so React is the one source of truth. One handler can serve every field through the input's `name`:

```tsx
const [values, setValues] = useState({ name: "", seats: "1" });
<input name="seats" value={values.seats} onChange={(e) => setValues({ ...values, [e.target.name]: e.target.value })} />
```

Inputs always hold strings, even `type="number"`. Convert when you send: `Number("")` is `0` and `Number("abc")` is `NaN`, so treat an empty box as `NaN` and let validation reject it.

**Validation on both sides.** The browser checks for speed: the passenger sees a mistake without a round trip. The server checks for safety: anyone can send any JSON with `curl`, so the browser's checks protect nothing. Keep the rules and messages identical, here by writing `validate.ts` to mirror `Validate` in `api/main.go`. `noValidate` on the `<form>` turns off the browser's built-in bubbles so your messages are the only ones.

**Showing errors.** `422 Unprocessable Content` with `{"errors": {"field": "message"}}` means "I understood you, but these fields are wrong". Put each message in an element with an `id`, and point its input at it with `aria-describedby`; add `aria-invalid="true"`. Screen readers then read the error with the field. Day 72 goes further.

**A preflight.** A `POST` with `Content-Type: application/json` isn't a "simple" request, so before sending it the browser asks with `OPTIONS`. The `cors` function in `api/main.go` answers with `Access-Control-Allow-Methods` and `Access-Control-Allow-Headers`. You can see both requests in the Network tab.

## Why it matters at work

"Validation only in the front end" is one of the most common findings in a security review: the API trusts whatever arrives (see [owasp-checklist.md](../../reference/owasp-checklist.md)). The opposite mistake is common too: rules that differ, so the browser accepts a phone number the server rejects with a message nobody sees. Keep rules in one place per side, test both with the same table, and always render the server's field errors.

## Steps

1. In `starter/validate.ts`, write `normalisePhone`: remove spaces and dashes with `replace(/[ -]/g, "")`, then match `/^0([17]\d{8})$/` or `/^\+?254([17]\d{8})$/` and return `"254" + group 1`, or `null`.
2. Write `validateBooking` with the rules in its comment. Count name characters with `[...name].length`, which matches Go's `utf8.RuneCountInString`; `name.length` counts an emoji as 2. Use `Number.isInteger` for seats. Run the tests: `validate.test.ts` should pass.
3. In `starter/App.tsx`, keep the four values as strings in one state object, plus `errors`, `sending`, the `booking` and a `failure` message.
4. Render each field with a `<label htmlFor>` and matching `id`. Route is a `<select>` with `Choose...` (value `""`) and then `ROUTE_NUMBERS`.
5. On submit: `preventDefault`, convert seats, run `validateBooking`. If there are errors, show them and stop. Otherwise set `sending`, `await createBooking(...)` (already written in `api.ts`), then show the booking or the server's errors. Catch anything thrown into `failure`, and clear `sending` in `finally`.
6. For each field with an error, render `<p id="FIELD-error">` and give the input `aria-invalid` and `aria-describedby`.
7. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-069
```

8. Try it against Go: `go run ./phase-6-frontend/day-069-forms/api` from the repo root, and `npm run dev -- day-069-forms/starter`. Book 4 seats on the 111 to see a server-only error. Then book with `curl`, skipping the browser's checks entirely; the server still says 422:

```bash
curl -i -X POST localhost:8080/api/bookings -d '{"name":"","phone":"1","route":"x","seats":99}'
```

## The kata: Phone Numbers

Clean up a list of phone numbers typed every possible way, using today's rules, and count the unique ones. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 69
```

## When you're stuck

- **`A component is changing an uncontrolled input to be controlled`**: a `value` started as `undefined`. Give every field a starting string, even `""`.
- **The browser shows its own "Please fill in this field" bubble**: add `noValidate` to the `<form>`.
- **"half a seat" or "an empty number box" passes**: `seats >= 1 && seats <= 14` is true for 1.5, and every comparison with `NaN` is false. Use `Number.isInteger(seats)` first.
- **"50 emoji count as 50" fails**: you used `name.length`. Use `[...name].length`.
- **The 422 message never appears**: `createBooking` returns `{ ok: false, errors }` for a 422 rather than throwing. Check `result.ok`.
- **The button stays on "Booking..." after an error**: reset `sending` in a `finally` block.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
