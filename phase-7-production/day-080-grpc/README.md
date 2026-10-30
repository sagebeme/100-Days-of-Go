# Day 80: gRPC

## The brief

Twende has split in two. The booking API that phones talk to stays JSON over HTTP, but the seat inventory is now its own service, and the booking API, the conductor's tablet app and the operations screen all call it. Between services, the team wants a strict contract that both sides compile against, fast binary messages, and a way to push seat changes to the operations screen as they happen. That's gRPC.

```
GetTrip   {trip_id: "46K-0730"}                           →  {id: "46K-0730", route: "46K Kencom to Kawangware", seats: 14, taken: [1, 3]}
BookSeat  {trip_id: "46K-0730", seat: 3, phone: "0712…"}  →  {id: "B1", seat: 3}
BookSeat  {trip_id: "46K-0730", seat: 3, phone: "0722…"}  →  error code AlreadyExists "seat 3 is taken"
BookSeat  {trip_id: "46K-0730", seat: 15, …}              →  error code InvalidArgument
GetTrip   {trip_id: "999-0000"}                           →  error code NotFound

WatchSeats {trip_id: "111-0800"}  →  stream: {seat: 0, seats_free: 2}   (now)
                                             {seat: 2, seats_free: 1}   (someone booked seat 2)
                                             {seat: 3, seats_free: 0}   …until the client cancels
```

*Route 46K (Citi Hoppa, Kencom to Kawangware) and 111 (MCA Sacco, Railways to Ngong) are real, checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026; the trip times and seat counts are made up. Routes change, so confirm with the conductor.*

## What you'll use

**Protocol Buffers.** `seatpb/seat.proto` is the contract. A `service` lists the calls; a `message` is a typed record. Each field has a number, and the number, not the name, is what goes over the wire.

```proto
service SeatService {
  rpc GetTrip(GetTripRequest) returns (Trip);
  rpc WatchSeats(WatchSeatsRequest) returns (stream SeatUpdate); // server streaming
}
message Trip {
  string id = 1;
  int32 seats = 4;
  repeated int32 taken = 5; // a list
}
```

Two rules keep old clients working when the contract grows: never reuse or change a field's number, and only add new fields (old code ignores fields it doesn't know). To retire a field, mark it `reserved 6;`.

**Generated code.** `protoc` turns the `.proto` into Go: `seat.pb.go` has a struct per message, and `seat_grpc.pb.go` has a `SeatServiceServer` interface to implement, a `SeatServiceClient` to call, and `RegisterSeatServiceServer`. The generated code is committed, so you don't need `protoc` to do today's work. Read it with `go doc ./phase-7-production/day-080-grpc/seatpb`. Use the getters, `req.GetTripId()`, which are safe on a nil message.

**Status codes, not HTTP codes.** A gRPC error carries a code from `google.golang.org/grpc/codes`. `status.Error(codes.NotFound, "no trip")` makes one; `status.Code(err)` reads it back on the client.

| Situation | Code | HTTP cousin |
|---|---|---|
| bad input | `InvalidArgument` | 400 |
| no such thing | `NotFound` | 404 |
| it already exists | `AlreadyExists` | 409 |
| no or bad credentials | `Unauthenticated` | 401 |
| not allowed | `PermissionDenied` | 403 |
| the caller's deadline passed | `DeadlineExceeded` | 504 |
| the caller cancelled | `Canceled` | 499 (nginx) |

**Deadlines.** A gRPC call carries its context's deadline to the server, which sees it in its own `ctx`. A client without a deadline can wait for ever on a stuck server, so `Client` wraps every call in `context.WithTimeout` ([Day 44](../../phase-4-concurrency/day-044-context-cancel/)).

**Server streaming.** The handler gets a `stream` and calls `stream.Send` as often as it likes; it ends the stream by returning. `stream.Context()` is cancelled when the client goes away. Today's server keeps a buffered channel per watcher, and `BookSeat` sends to each one with a `select` and `default`, so one slow watcher can't block a booking ([Day 43](../../phase-4-concurrency/day-043-select-timeouts/)).

**bufconn.** `google.golang.org/grpc/test/bufconn` is a listener that lives in memory. The tests run a real gRPC server and client over it, with no ports and no network.

**The package name.** The folder is `day-080-grpc`, so the package is `grpc`, and it imports `google.golang.org/grpc`, whose name is also `grpc`. That's allowed: inside the files, `grpc.` means the import. The runnable server in `cmd/seats` imports it as `seats` to keep things readable.

## Why it matters at work

Most companies with more than a handful of services talk gRPC (or something very like it) between them, and the `.proto` file becomes the most reviewed file in the repo. Reviewers check that field numbers never change, that errors use the right status code (clients retry `Unavailable` but not `InvalidArgument`), and that every client call has a deadline, because one service with no timeouts can hang every service that calls it. Streaming handlers are where goroutines leak: a handler that doesn't watch `stream.Context()` keeps running after the client has gone. Your test for "the watcher is removed after cancel" is the kind of test that catches that.

## Steps

1. Read `seatpb/seat.proto`, then skim `seatpb/seat_grpc.pb.go` for the `SeatServiceServer` interface. You implement those three methods.
2. Open `starter/seats.go`. `Server`, `NewServer`, the errors, `Client` and `NewClient` are given.
3. Write `GetTrip`. Empty ID: `InvalidArgument`. Under the lock, an unknown trip is `NotFound`. Build `Taken` from the map's keys and sort it with `slices.Sort`.
4. Write `BookSeat`. Check the inputs, then under the lock: `NotFound`, `InvalidArgument` for a seat outside `1..seats`, `AlreadyExists` for a taken one. Record the booking, count it for the `B1`, `B2` ID, and send a `SeatUpdate` to each watcher of that trip without blocking.
5. Write `WatchSeats`. Under the lock, look up the trip (`NotFound`), work out seats free, and register a buffered channel in `s.watchers[tripID]`. Unlock. `defer` removing the channel. Send the first update with seat 0, then loop: `select` on `stream.Context().Done()` (return nil) and the channel (`stream.Send`).
6. Write `translate`, then `SeatsFree` and `Book` on `Client`: a `context.WithTimeout` for each call, the generated client, and `translate` on errors.
7. Run the tests from the repo root:

```bash
go test -race ./phase-7-production/day-080-grpc/starter
```

8. Run the real server and call it. Install [grpcurl](https://github.com/fullstorydev/grpcurl), then:

```bash
go run ./phase-7-production/day-080-grpc/starter/cmd/seats
grpcurl -plaintext -d '{"trip_id":"46K-0730"}' localhost:50051 twende.seats.v1.SeatService/GetTrip
```

**Changing the contract.** If you edit `seat.proto`, regenerate the Go code. Install the two Go plugins, and `protoc` itself from the [protobuf releases page](https://github.com/protocolbuffers/protobuf/releases) (unzip it and put its `bin` on your `PATH`). Then from the repo root:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
protoc -I phase-7-production/day-080-grpc/seatpb \
  --go_out=. --go_opt=module=github.com/sagebeme/100-Days-of-Go \
  --go-grpc_out=. --go-grpc_opt=module=github.com/sagebeme/100-Days-of-Go \
  seat.proto
```

The committed files were made with protoc 33.1 and those two plugin versions.

## The kata: Top Routes

Keep running seat totals per route from a stream of bookings and answer "top k" queries with a min-heap. It's the classic top-k interview question, and what the operations screen fed by `WatchSeats` would show. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 80
```

## When you're stuck

- **`code = Unimplemented desc = GetTrip is not written yet`**: that method still has the starter body.
- **`GetTrip("") code = NotFound, want InvalidArgument`**: check for an empty ID before looking the trip up.
- **`code = Unknown`** where you wanted another code: you returned a plain `errors.New` or `fmt.Errorf`. Use `status.Error(codes.X, "...")` or `status.Errorf`.
- **`Taken = [3 1]`**: map order is random. Sort the slice.
- **`TestWatchSeats` hangs until the 2 s deadline, then fails on `Recv`**: the update never reached the stream. Check you register the channel under the trip's ID, that `BookSeat` sends to the watchers of `t.id`, and that the first message is sent.
- **`1 watchers still registered after the client cancelled`**: the `defer` that deletes the channel is missing, or your loop never looks at `stream.Context().Done()`.
- **`SeatsFree against a hung server = <nil>`** or the test takes seconds: the client call has no timeout. Wrap it in `context.WithTimeout(ctx, c.timeout)` and `defer cancel()`.
- **`Book(taken) = rpc error: code = AlreadyExists`, want `ErrSeatTaken`**: pass client errors through `translate`.
- **Still stuck?** Open `solution/seats.go`, read it once, close it, then write your own version from memory.
