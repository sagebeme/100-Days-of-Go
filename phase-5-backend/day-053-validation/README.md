# Day 53: Validation

## The brief

The bookings table has filled up with rubbish. Seat 0. Seat 99 on a 14-seater. A passenger called `<script>alert(1)</script>`. Phone numbers typed four different ways, so one passenger looks like three. A booking for the 07:30 to Rongai made at 09:00. And once, a 40 MB "passenger name" that took the server down for a minute.

Today every booking is checked at the door, before it touches the store. The body is capped in size, each field gets a clear rule, phones are normalised to one form, and the client hears about every problem in one answer, so the app can mark each field on the form at once.

```
POST /trips/1/bookings  {"seat": 3, "passenger": " Achieng ", "phone": "0712 345 678"}
  →  201  {"id":1,"tripId":1,"seat":3,"passenger":"Achieng","phone":"254712345678"}

POST /trips/2/bookings  {"seat": 0, "passenger": "", "phone": "0712"}
  →  422  {"error":"validation failed","fields":[
             {"field":"passenger","message":"is required"},
             {"field":"phone","message":"must be a Kenyan mobile number like 0712 345 678"},
             {"field":"seat","message":"must be between 1 and 14"},
             {"field":"trip","message":"has already left"}]}

POST /trips/1/bookings  (a 5 KB body)  →  413  {"error":"body must be 4096 bytes or less"}
POST /trips/1/bookings  {"seat": 3     →  400  {"error":"body must be valid JSON with seat, passenger and phone"}
```

## What you'll use

- **`http.MaxBytesReader(w, r.Body, n)`** wraps the body so reading past `n` bytes fails with a `*http.MaxBytesError`. It also tells the server to close the connection afterwards. Without it, `json.NewDecoder(r.Body)` reads whatever the client sends.
- **Three stages, three codes.** Can I read it? (`413` too big, `400` broken JSON.) Does the thing it's about exist? (`404`.) Do the values follow the rules? (`422`, with every problem.) [reference/http.md](../../reference/http.md) has the "400 or 422?" rule.
- **A slice of field errors that is an `error`.** `type Errors []FieldError` with an `Error()` method can be returned and logged like any error, and encodes as a JSON list.
- **Runes, not bytes, for length.** `len("Wanjirũ")` is 8 bytes, but it's 7 characters. `utf8.RuneCountInString` counts characters. `unicode.IsLetter` accepts letters from any alphabet, so a name isn't rejected for having a `ũ` in it.
- **Normalise, then store.** The phone check is your [Day 9](../../phase-1-terminal/day-009-phone-cleaner/) `Clean`. Store the clean form, `254712345678`, so the same passenger always looks the same.
- **A clock you pass in.** `NewAPI(store, now)` takes `now func() time.Time`, so a test can fix the time and "has already left" doesn't depend on when the tests run.

```go
r.Body = http.MaxBytesReader(w, r.Body, MaxBody)
if err := dec.Decode(&req); err != nil {
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		// 413
	}
}

utf8.RuneCountInString("Wanjirũ") // 7
unicode.IsLetter('ũ')             // true
```

## Why it matters at work

Validation is the cheapest security you'll ever add. OWASP's injection and insecure-design items both start with "don't trust input" (see [reference/owasp-checklist.md](../../reference/owasp-checklist.md)), and a body size limit is the one line that stops a single client from filling your server's memory. Allow-lists, like "letters, spaces, hyphens, apostrophes and full stops", are safer than block-lists, because you can't list everything bad. In review, people check three things: the body is limited before it's read, every field the handler uses has a rule, and the errors tell a front-end developer which field to fix. Validation doesn't replace escaping output or parameterised SQL ([Day 55](../day-055-postgres/)). It's a first wall, not the only one.

## Steps

`starter/store.go` is written for you. You write `starter/validate.go` and `starter/api.go`.

1. Write `Errors.Error`: each problem as `field message`, joined with `"; "`.
2. Write `NormalizePhone`. Copy your Day 9 `Clean` and adjust the error messages if you like.
3. Write `ValidateBooking`. Start with `var errs Errors` and a small `add(field, msg)` closure that appends to it. Check the fields in the order of the doc comment, giving each field at most one message:
   - Trim the passenger. Then, in a `switch`: empty is "is required"; fewer than 2 or more than 60 runes is the length message; any rune that isn't a letter, space, `-`, `'` or `.` is the characters message.
   - A blank phone is "is required". Otherwise `NormalizePhone` it, and on an error add the phone message. On success, keep the normalised number in `req.Phone`.
   - The seat must be from 1 to `trip.Seats`.
   - `trip.Departs.After(now)` must be true.
   - Return the cleaned `req` and `errs`. If nothing was added, `errs` is still nil.
4. In `api.go`, write the handler for `POST /trips/{id}/bookings`. First line: `r.Body = http.MaxBytesReader(w, r.Body, MaxBody)`. Decode with `DisallowUnknownFields`. A `*http.MaxBytesError` is `413`, any other decode error is `400`.
5. Parse the trip ID (`400`) and load the trip (`404` for `ErrNotFound`).
6. Call `ValidateBooking` with `a.now()`. If there are errors, answer `422` with `{"error": "validation failed", "fields": errs}`. Make a small struct for that body.
7. Book with the **cleaned** request. `ErrSeatTaken` is `409`. Otherwise set `Location` and answer `201`.
8. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-053-validation/starter
```

## The kata: IPv4 Check

Client IP addresses feed rate limits and logs, and some arrive in headers like `X-Forwarded-For` that a client can fill with anything. The kata validates them by hand: four parts, digits only, no leading zeros, nothing over 255. It's also a classic interview question, because the edge cases are the whole problem. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 53
```

## When you're stuck

- **The 5 KB body gets `400`, not `413`**: you checked for the size error after giving up on it. Use `errors.As(err, &tooBig)` with `var tooBig *http.MaxBytesError`, before the generic `400`.
- **The 5 KB body is accepted**: `MaxBytesReader` must wrap `r.Body` before the decoder reads it. Make it the first line of the handler.
- **`"Wanjirũ Ng'ang'a"` is rejected**: you checked bytes with `c >= 'a' && c <= 'z'`. Range over the string to get runes, and use `unicode.IsLetter`.
- **60 `ũ` letters count as too long**: `len` counts bytes, and `ũ` is 2 bytes. Use `utf8.RuneCountInString`.
- **The test wants `errs` to be nil, and gets `Errors{}`**: you made an empty slice with `Errors{}`. Declare it with `var errs Errors` so it stays nil until something is added.
- **A field shows up twice in the errors**: use a `switch` or `else if` so each field stops at its first problem.
- **The stored phone is `"0712 345 678"`**: you booked with `req` instead of the cleaned request that `ValidateBooking` returned.
- **Still stuck?** Open `solution/validate.go` and `solution/api.go`, read them once, close them, then write your own version from memory.
