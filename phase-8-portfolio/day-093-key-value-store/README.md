# Day 93: Key-Value Store

## The brief

Achieng runs ChamaBox (a made-up app) for twelve chamas in Kisumu. It runs on one small box in her shop, and the power goes off a few times a week. Last month a blackout hit in the middle of a save, the app's JSON file came back half-written, and a chama lost its contribution records. She wants a tiny database that never loses a write it said was saved, and that starts up cleanly after the power comes back, even if the last write was cut off halfway.

```
s, _ := Open("data")
s.Set("chama:balance", "Ksh 45,000")       →  nil   (on disk before it returns)
s.Get("chama:balance")                     →  "Ksh 45,000", true
s.Delete("chama:balance")                  →  nil
s.Get("chama:balance")                     →  "", false

// power cut while writing the last record, then:
s, _ = Open("data")                        →  every earlier record back; the torn one dropped
s.Size()                                   →  1204
s.Compact()                                →  nil
s.Size()                                   →  41  (same keys and values)
```

**Acceptance criteria**

- Every `Set` and `Delete` is appended to a log file and synced to disk before it returns.
- `Open` replays the log. A torn or corrupt last record is dropped, never read as a wrong value, and the next write still survives a restart.
- `Compact` shrinks the log to one record per live key, and a crash during `Compact` loses nothing.
- The store is safe for many goroutines at once.

## What you'll use

- Files, `os.OpenFile` and syncing to disk, from [Day 23: Save and Load](../../phase-2-modelling/day-023-save-load/).
- Sentinel errors and `errors.Is`, from [Day 17: Failed Transactions](../../phase-2-modelling/day-017-failed-transactions/).
- `sync.RWMutex` and the race detector, from [Day 45: Mutexes](../../phase-4-concurrency/day-045-mutexes/).
- Table-driven tests and helpers, from [Day 19: Table-Driven Tests](../../phase-2-modelling/day-019-table-tests/).
- New today: `encoding/binary` for fixed-size integers, `hash/crc32` for a checksum, and `t.TempDir()` for a fresh folder per test that Go deletes afterwards.

## Steps

Requirements. The tests in `starter/store_test.go` check every must-have.

**Must-haves**

1. `Open(dir)` creates the folder and `dir/data.wal` if they're missing, and loads the keys by replaying the log from the start.
2. Each record says what changed: set or delete, the key, and for a set the value. A suggested binary layout, all integers big-endian:

   ```
   crc32 (4 bytes) | op (1) | key length (4) | value length (4) | key | value
   ```

   The CRC-32 covers everything after itself. Lengths make the format safe for any bytes in a key or value. A text format like `SET key value\n` breaks on a value with a space or a newline in it.
3. `Set` and `Delete` write the record, call `f.Sync()`, and only then change the map. If the write fails, the map stays as it was.
4. `Open` stops at the first record that is short or has a bad checksum, cuts the file there with `f.Truncate`, and carries on from that point. A length over `MaxKeySize` or `MaxValueSize` also means a bad record.
5. `Delete` of a missing key returns `nil` and writes nothing. An empty key is `ErrEmptyKey`. A key or value over its limit is `ErrTooLarge`.
6. `Compact` writes the live keys to `data.wal.tmp`, syncs it, then renames it over `data.wal`. Afterwards, only `data.wal` is in the folder.
7. After `Close`, every write returns `ErrClosed`, and so does a second `Close`.
8. Run the tests, then run them again with the race detector:

```bash
go test ./phase-8-portfolio/day-093-key-value-store/starter
go test -race ./phase-8-portfolio/day-093-key-value-store/starter
```

**Stretch goals**

- Compact by itself when the log is, say, four times bigger than the live data.
- A `starter/cmd/kv` command: `kv set k v`, `kv get k`, `kv compact`, with exit codes as on Day 22.
- A `GetAt(key, t)` that answers "what was the value at time t?". That's the kata.
- Batch writes: `Apply(ops)` writes several records and syncs once. Fewer syncs is the biggest speed-up a WAL has.

## The kata: Time Travel Store

Process `SET`, `DEL` and `GET` commands with timestamps, and answer each `GET` with the value the key had at that time. It's a common interview question (LeetCode calls it "time-based key-value store"), and it is your log kept as a sorted history per key, searched with binary search from [Day 32](../../phase-3-algorithms/day-032-binary-search/). Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 93
```

## When you're stuck

- **`unexpected EOF` from `io.ReadFull` on restart**: that's the torn record. Treat `io.ErrUnexpectedEOF` as "stop here", and plain `io.EOF` before any byte as "the log ended cleanly".
- **The crash test passes once, then the next write disappears after a restart**: you skipped the bad bytes in memory but left them in the file. New records land after the garbage and the next `Open` stops before them. Truncate the file to the last good offset in `Open`.
- **"last byte flipped" returns `charlie` with a wrong letter**: you need the checksum. Compute `crc32.ChecksumIEEE` over the record when writing and compare it when reading.
- **`Open` tries to allocate a huge slice on a damaged file**: a flipped byte in a length field. Check lengths against the limits before `make`.
- **Writes after `Compact` vanish after a restart**: you kept writing to the old file handle, which now points at a deleted file. After the rename, write to the new file.
- **The race detector reports `Get` against `Set`**: reads touch the map too. Take `RLock` in `Get`, `Len`, `Keys` and `Size`.
- **Still stuck?** Open `solution/store.go`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- **README**: what it is in one line, the record format as a diagram, the guarantees (what `Set` returning `nil` promises) and the limits (one process, keys in memory, values up to 1 MiB).
- **Demo**: a short recording of the stretch `kv` command. Write keys, pull the plug with `kill -9` in the middle of a loop of writes, restart, show the keys are back.
- **Tests and CI**: run `go test -race ./...` in GitHub Actions (Day 63) and put the badge at the top of the README. Add a fuzz test (`go test -fuzz`) that feeds random bytes to `Open` and checks it never panics.
- **Deploy link**: this is a library, so the "deploy" is a tagged release (`v0.1.0`) and a [pkg.go.dev](https://pkg.go.dev) page once someone runs `go get` on it.
- **What you'd do next**: write it up honestly. Keys live in memory, so it can't hold more than RAM. Bitcask keeps only the keys and file offsets in memory; LSM trees (LevelDB, Pebble) sort data on disk. A bad record in the middle of the log currently drops everything after it; a real store would report it.
