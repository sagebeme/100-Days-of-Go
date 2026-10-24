# Day 15: Boda Boda Fleet

## The brief

Mzee Kariuki owns three boda bodas in Githurai. Every evening the riders tell him how far they went and what they collected, and he writes it in an exercise book. The book has coffee stains, and last month a bike went 1,400 km without a service because nobody added up the numbers.

Build him a fleet tracker. Each bike knows its plate, its rider, its odometer, its trips and its earnings. The fleet records trips, adds up the money and says which bikes are due for a service.

```
b := Bike{Plate: "KMFA 123B", Rider: "Otieno", Odometer: 990}
b.Label()                       →  "KMFA 123B (Otieno)"
b.Ride(15, 300)                 →  nil   (b.Odometer is now 1005, b.Earnings 300, b.Trips 1)
b.NeedsService()                →  true  (1005 km since the last service at 0)
b.Ride(0, 100)                  →  error: a trip must be at least 1 km

f := &Fleet{}
f.Add(b)                        →  nil
f.Add(Bike{Plate: "KMFA 123B"}) →  error: KMFA 123B is already in the fleet
f.Ride("KMFA 123B", 4, 150)     →  nil
f.TotalEarnings()               →  450
f.DueForService()               →  ["KMFA 123B"]
```

## What you'll use

New today: **structs** and **methods**.

- A **struct** groups named fields into one type. `Bike{Plate: "KMFA 123B", Rider: "Otieno"}` makes one; fields you leave out get their zero value (`0`, `""`, `false`).
- A **method** is a function with a receiver, written before the name: `func (b Bike) Label() string`. You call it as `b.Label()`.
- A **value receiver** `(b Bike)` gets a copy of the bike. Use it for methods that only read.
- A **pointer receiver** `(b *Bike)` gets the address of the caller's bike, so changes stick. Use it for methods that change the struct. Go takes the address for you: if `b` is a variable, `b.Ride(4, 150)` means `(&b).Ride(4, 150)`.
- `&x` gives a pointer to `x`. `nil` is the zero value of a pointer: it points at nothing.
- Structs are compared with `==` field by field, as long as every field can be compared.

```go
type Matatu struct {
	Route string
	Seats int
}

func (m Matatu) Full(passengers int) bool { return passengers >= m.Seats } // reads only
func (m *Matatu) AddSeats(n int)          { m.Seats += n }                 // changes m

m := Matatu{Route: "46", Seats: 14}
m.AddSeats(2) // m.Seats is 16
```

The trap that catches everyone: `range` gives you a copy of each element.

```go
for _, b := range f.Bikes {
	b.Rider = "Njoroge" // changes the copy; the fleet doesn't change
}
for i := range f.Bikes {
	f.Bikes[i].Rider = "Njoroge" // changes the bike in the slice
}
```

`Find` uses the same idea. `return &f.Bikes[i]` returns a pointer to the bike inside the slice, so whoever calls `Find` can change it.

## Steps

1. Open `starter/fleet.go`. The types are written. Read the field names and the doc comments; the tests use them exactly.
2. Write `Label`. When `Rider` is empty, use `no rider`. `fmt.Sprintf("%s (%s)", ...)` builds it.
3. Write `Ride`. Check `km` and `fare` first and return an error before you change anything. Then add to `b.Odometer` and `b.Earnings`, and `b.Trips++`.
4. Write `NeedsService` and `Service`. The distance since the last service is `b.Odometer - b.LastServiceKm`.
5. Try this once: change `Ride`'s receiver from `(b *Bike)` to `(b Bike)` and run the tests. `TestRide` fails because `Ride` now changes a copy. Change it back.
6. Write `Find`. Loop with `for i := range f.Bikes` and return `&f.Bikes[i]` when the plate matches. Return `nil` at the end.
7. Write `Add` using `Find` to spot a duplicate. Append with `f.Bikes = append(f.Bikes, b)`. That's why `Add` has a pointer receiver: `append` may give back a new slice, and the fleet must keep it.
8. Write `Fleet.Ride`: find the bike, return an error if it's `nil`, otherwise call the bike's `Ride` and return what it returns.
9. Write `TotalEarnings` and `DueForService`. Start `DueForService` with `due := []string{}` so an empty fleet gives an empty slice, not `nil`, and sort it with `sort.Strings`.
10. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-015-boda-fleet/starter
```

## The kata: Rider Leaderboard

A list of trips comes in, one line per trip. Add them up per rider with a struct, sort the riders with `sort.Slice` and a three-level tie-break, and print the top `k`. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 15
```

## When you're stuck

- **`TestRide` says the odometer didn't change**: your method has a value receiver `(b Bike)`, so it changed a copy. Use `(b *Bike)`.
- **`TestFleetFind` says the fleet's rider didn't change**: you returned a pointer to a copy. `for _, b := range f.Bikes { return &b }` points at the loop variable, not the slice element. Return `&f.Bikes[i]`.
- **`f.Bikes` is still empty after `Add`**: `Add` needs the pointer receiver `(f *Fleet)`. With `(f Fleet)` the append happens to a copy.
- **`invalid memory address or nil pointer dereference`**: you used the result of `Find` without checking it for `nil`.
- **`DueForService() = []string(nil), want []string{}`**: start with `[]string{}`, not `var due []string`.
- **`Ride` changed the bike and still returned an error**: do every check first, then change the fields.
- **Still stuck?** Open `solution/fleet.go`, read it once, close it, then write your own version from memory.
