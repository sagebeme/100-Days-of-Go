# Day 23: Save and Load

## The brief

Every chama tool so far forgets everything when it exits. The Umoja Women's Group wants its contributions kept: who paid, how much and when. Give the ledger a file. `Save` writes it as JSON that a person can read, `Load` reads it back and complains clearly about a broken file, and nothing secret (like a member's PIN) ever reaches the disk.

```
l := Ledger{Chama: "Umoja Women's Group", Members: []Member{{Name: "Wanjiku", Phone: "0712345678", PIN: "4321"}}}
l.Contribute("Wanjiku", "2026-03-01", 500)  →  nil
l.Contribute("Wanjiku", "2026-02-30", 500)  →  error: date "2026-02-30": want YYYY-MM-DD
Save("ledger.json", l)                      →  nil
```

`ledger.json` now holds:

```json
{
  "chama": "Umoja Women's Group",
  "members": [
    {
      "name": "Wanjiku",
      "phone": "0712345678",
      "contributions": [
        {
          "date": "2026-03-01",
          "amount_ksh": 500
        }
      ]
    }
  ]
}
```

```
Load("ledger.json")      →  the same ledger, without the PIN
Load("missing.json")     →  error wrapping fs.ErrNotExist
LoadOrNew("missing.json", "Umoja")  →  a new empty ledger, nil
```

## What you'll use

New today: **files**, **`encoding/json`** and **struct tags**.

- **Struct tags** are strings after a field that packages can read. `encoding/json` uses the `json` tag to pick the name in the file:

```go
type Contribution struct {
	Date   string `json:"date"`
	Amount int    `json:"amount_ksh"`
}
```

  - `json:"notes,omitempty"` leaves the field out when it's empty.
  - `json:"-"` never writes or reads the field. Use it for anything that must not be saved.
  - Only **exported** (capitalised) fields are saved. A lower-case field is silently skipped.
- `json.MarshalIndent(v, "", "  ")` turns a value into indented JSON bytes. A `nil` slice becomes `null` and an empty one becomes `[]`.
- `json.NewDecoder(r).Decode(&v)` reads JSON into `v`. Call `dec.DisallowUnknownFields()` first and a typo like `"ammount_ksh"` becomes an error, instead of a field that quietly stays 0.
- `os.ReadFile(path)` reads a whole file. Its error for a missing file wraps `fs.ErrNotExist`, so `errors.Is(err, fs.ErrNotExist)` works (Day 17), even after you wrap it again with `%w`.
- **Writing safely**: `os.CreateTemp(dir, pattern)` makes a new file with a unique name, readable only by you. Write to it, `Close` it, then `os.Rename` it over the real file. On Linux and macOS a rename within one folder is atomic: anyone reading the file sees the old version or the new one, never half of each. If the program crashes halfway, the old ledger is still there.
- `defer os.Remove(tmp.Name())` cleans up the temporary file if anything fails. Once the rename has happened, there's nothing left to remove, and the error is ignored.
- `time.Parse(time.DateOnly, date)` checks a `YYYY-MM-DD` date. `time.DateOnly` is the layout `"2006-01-02"`. Go layouts are written with that exact reference date, not letters like `YYYY`.

## Why it matters at work

Files are where data gets lost. In review, look for the write-then-rename pattern. Writing straight over the real file is fine until the disk fills up or the laptop dies halfway, and then the ledger is gone. Ask what the file permissions are. Phone numbers and payment records are personal data under Kenya's Data Protection Act, so `0644` (anyone on the machine can read it) is the wrong default. Look for tags on every field, because renaming a Go field without one silently changes the file format, and old files stop loading. And look for `DisallowUnknownFields` wherever a person might edit the file by hand. Silent typos in config and data files cause real outages.

## Steps

1. Open `starter/ledger.go`. Add `json` tags to every field of the three structs, using the names in the example above. `Notes` gets `omitempty`. `PIN` gets `-`.
2. Write `Contribute`. Check the amount and the date with `time.Parse(time.DateOnly, date)` first. Then find the member with `for i := range l.Members` and append to `l.Members[i].Contributions`. (Day 15: a `range` value is a copy.)
3. Write `Total`.
4. Write `Save`. Turn `nil` slices into empty ones so the file says `[]`, not `null`. Do that on a copy of `Members`, not the caller's slice. Then `json.MarshalIndent`, add a `'\n'`, and write it with `os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")`, `Write`, `Close`, `os.Chmod(name, 0o600)` and `os.Rename`. Wrap every error with the path.
5. Write `Load` with `os.ReadFile` and a `json.Decoder` with `DisallowUnknownFields`. Wrap errors with `%w`, so a missing file still matches `fs.ErrNotExist`.
6. Write `LoadOrNew` with `errors.Is(err, fs.ErrNotExist)`.
7. Run the tests from the repo root. They write only inside `t.TempDir()`, a fresh folder per test that Go deletes afterwards:

```bash
go test ./phase-2-modelling/day-023-save-load/starter
```

## The kata: Till Totals

Decode a JSON export of Lipa na M-Pesa payments with tagged structs, skip the reversed ones and total the rest per till. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 23
```

## When you're stuck

- **The JSON has `"Name"` and `"Amount"`**: a tag is missing or misspelled. Tags use backquotes and no spaces: `` `json:"name"` ``. Run `go vet`; it catches many broken tags.
- **The file has `"PIN": "4321"`**: tag it `` `json:"-"` ``.
- **The file has `"contributions": null`**: the slice was `nil`. Replace it with `[]Contribution{}` before you marshal.
- **`TestSaveDoesNotChangeTheCaller` fails**: you wrote into `l.Members[i]` directly. The struct `l` is a copy, but its `Members` slice shares the caller's array. Make a new slice and fill that.
- **`TestSaveLeavesOneFile` finds `ledger.json.tmp-123456`**: the rename didn't happen, or you didn't remove the temporary file after an error. `defer os.Remove(tmp.Name())` right after `CreateTemp`.
- **`invalid cross-device link`**: you made the temporary file in `os.TempDir()`. A rename only works within one filesystem, so create it in the same folder as `path`.
- **`TestLoadMissing` fails though `Load` returns an error**: you wrapped it with `%v`. Use `%w`.
- **`parsing time "2026-02-30": day out of range`**: that's `time.Parse` doing its job. Return your own error message.
- **Still stuck?** Open `solution/ledger.go`, read it once, close it, then write your own version from memory.
