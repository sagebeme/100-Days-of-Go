# Day 52: JSON API

## The brief

The booking app now has two clients: Nduta's web front end, and a USSD gateway a partner Sacco wrote in a hurry. The gateway sends `"seats": "14"` as a string, sends `routeId` as `route_id`, and sends times with no time zone. Yesterday's API ignored `route_id` without a word, and every other mistake came back as the same "body must be valid JSON", which told the gateway's developer nothing. Meanwhile the web front end showed a trip at 04:30, because the API sent times in UTC.

Today the API gets strict, and helpful about it. It only accepts JSON, rejects fields it doesn't know, and says exactly what was wrong. Every time it sends carries Nairobi's offset, so nobody has to guess.

```
POST /trips  Content-Type: application/json
{"routeId": 125, "departs": "2026-11-02T04:30:00Z", "seats": 14}
  →  201  Location: /trips/1
          {"id":1,"routeId":125,"departs":"2026-11-02T07:30:00+03:00","seats":14}

POST /trips  Content-Type: text/plain             →  415  {"error":"Content-Type must be application/json"}
{"route_id": 125, ...}                            →  400  {"error":"body has unknown field \"route_id\""}
{"seats": "14", ...}                              →  400  {"error":"field \"seats\" must be a whole number"}
{"departs": "2026-11-02T07:30:00", ...}           →  400  {"error":"times must look like \"2026-11-02T07:30:00+03:00\""}
{"routeId": 125, ...}{"routeId": 125, ...}        →  400  {"error":"body must hold a single JSON value"}

GET /trips?date=2026-11-02                        →  200  [trips leaving that day in Nairobi]
GET /trips?date=02-11-2026                        →  400  {"error":"date must look like \"2026-11-02\""}
```

## What you'll use

- **`json.NewDecoder(r.Body)`** reads JSON straight from the body. `dec.DisallowUnknownFields()` turns a field your struct doesn't have into an error instead of ignoring it, which catches mistakes like `route_id`. Matching is case-insensitive, though: `routeID` still fills `routeId`.
- **One value per body.** `Decode` stops after the first JSON value. Call it again: anything but `io.EOF` means there was more.
- **Decoding errors have types.** `*json.SyntaxError` has an `Offset`, `*json.UnmarshalTypeError` has the `Field` and the Go `Type` it wanted, a bad time is a `*time.ParseError`, and an empty body is `io.EOF`. `errors.As` picks them apart, so each gets its own message. Unknown fields have no type of their own, only the message `json: unknown field "x"`.
- **`mime.ParseMediaType`** (new today) splits `application/json; charset=utf-8` into the media type and its parameters, so you don't compare the whole header with `==`.
- **`reflect.Type.Kind()`** (new today, and only here) tells you whether the wanted type was an `int` or a `string`, so the message can say "a whole number" instead of Go's type name.
- **Times.** `time.Time` encodes as RFC 3339, `2026-11-02T07:30:00+03:00`, in whatever zone it's in. Store UTC, send local time with its offset, and accept any offset on the way in. Kenya is UTC+3 all year with no daylight saving, so `time.FixedZone("EAT", 3*60*60)` is exact. `time.ParseInLocation("2006-01-02", s, Nairobi)` reads a date as midnight in Nairobi. Go's layouts are written with the reference time, 2006-01-02 15:04:05.

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
err := dec.Decode(&req)

var wrongType *json.UnmarshalTypeError
if errors.As(err, &wrongType) {
	fmt.Println(wrongType.Field, wrongType.Type.Kind()) // seats int
}
```

The JSON sections of [reference/http.md](../../reference/http.md) cover the helpers and the safe-reading pattern. Day 53 adds the body size limit and checks the values themselves.

## Why it matters at work

A JSON API is a contract with clients you don't control. If it quietly ignores fields and coerces values, every client bug becomes your data bug: zero-seat trips, wrong departure times, a `fare` field someone thinks they're setting. Rejecting unknown fields and wrong types with a clear message makes integrators fix their side in minutes instead of filing tickets for weeks. Time zones are the classic production incident. A server in a Frankfurt data centre, a database in UTC and passengers in Nairobi will show the wrong time somewhere unless every time on the wire carries its offset. In review, people ask two questions about any time field: what zone is it stored in, and what zone is it sent in?

## Steps

`starter/store.go` is written for you: the `Trip` type, `Nairobi`, a `MemStore` that saves times in UTC, and `writeJSON` and `writeError`.

1. Finish `RequestError.Error`: return the message.
2. Write `DecodeJSON`. Parse the `Content-Type` header with `mime.ParseMediaType`. If that fails or the type isn't `application/json`, return a `*RequestError` with status 415.
3. Make a decoder, call `DisallowUnknownFields`, and decode into `dst`. On an error, turn it into a 400 `*RequestError` with the messages from the doc comment. Check in this order: `io.EOF`, `io.ErrUnexpectedEOF`, `*json.SyntaxError`, `*json.UnmarshalTypeError` (an empty `Field` means the whole body was the wrong type, such as a list), `*time.ParseError`, then the `json: unknown field ` prefix.
4. Decode a second time into `&struct{}{}`. If that error isn't `io.EOF`, return "body must hold a single JSON value".
5. Write `ToJSON`: copy the fields, with `t.Departs.In(Nairobi)`.
6. `POST /trips`: `DecodeJSON` into a request struct. On a `*RequestError`, answer with its status and message. If `routeId` or `seats` is below 1 or `departs` is zero, answer `422`. Otherwise save, set `Location`, answer `201` with `ToJSON`.
7. `GET /trips`: build a `[]TripJSON` with `make(..., 0, n)` so it's `[]` when empty. With `?date=`, read it from `r.URL.Query().Get("date")`, parse it with `time.ParseInLocation`, and keep trips where `!t.Departs.Before(day) && t.Departs.Before(day.AddDate(0, 0, 1))`. A date that doesn't parse is `400`.
8. `GET /trips/{id}`: `200` with `ToJSON`, `400` or `404`.
9. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-052-json-api/starter
```

## The kata: JSON Brackets

Every JSON parser starts by checking that brackets close in the right order, skipping anything inside strings. The kata is that check, with a stack and the escape rules for `\"`, and it prints the position of the first problem the way a parser's error does. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 52
```

## When you're stuck

- **`json: unknown field "seat"` reaches the client as is**: that's the decoder's message, not yours. Match the prefix with `strings.HasPrefix` and build your own message.
- **Junk after the JSON is accepted**: `Decode` reads one value and stops. Call it a second time and expect `io.EOF`.
- **The empty-body test gets "malformed"**: an empty body gives `io.EOF`, and a body cut off half way gives `io.ErrUnexpectedEOF`. Check `io.EOF` first, with `errors.Is`.
- **`"application/json; charset=utf-8"` gets 415**: you compared the whole header. Use the media type that `mime.ParseMediaType` returns.
- **The response says `"2026-11-02T04:30:00Z"`**: you sent the stored UTC time. Convert with `.In(Nairobi)` in `ToJSON`.
- **The trip at 00:30 on 3 November shows up for 2 November**: you compared dates in UTC, where it's still the 2nd. Parse the date with `time.ParseInLocation` in `Nairobi` and compare instants with `Before`.
- **`GET /trips?date=...` with no trips returns `null`**: a nil slice encodes as `null`. Make it with `make([]TripJSON, 0, len(all))`.
- **Still stuck?** Open `solution/api.go`, read it once, close it, then write your own version from memory.
