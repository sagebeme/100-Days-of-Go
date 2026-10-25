package saveload

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func sample() Ledger {
	return Ledger{
		Chama: "Umoja Women's Group",
		Members: []Member{
			{
				Name:  "Wanjiku",
				Phone: "0712345678",
				Contributions: []Contribution{
					{Date: "2026-03-01", Amount: 500},
					{Date: "2026-04-01", Amount: 750},
				},
				Notes: "treasurer",
				PIN:   "4321",
			},
			{Name: "Achieng", Phone: "0798765432", PIN: "1111"},
		},
	}
}

const sampleJSON = `{
  "chama": "Umoja Women's Group",
  "members": [
    {
      "name": "Wanjiku",
      "phone": "0712345678",
      "contributions": [
        {
          "date": "2026-03-01",
          "amount_ksh": 500
        },
        {
          "date": "2026-04-01",
          "amount_ksh": 750
        }
      ],
      "notes": "treasurer"
    },
    {
      "name": "Achieng",
      "phone": "0798765432",
      "contributions": []
    }
  ]
}
`

func TestContribute(t *testing.T) {
	l := sample()
	if err := l.Contribute("Achieng", "2026-04-02", 1000); err != nil {
		t.Fatalf("Contribute(Achieng, 2026-04-02, 1000) returned an error: %v", err)
	}
	if err := l.Contribute("Achieng", "2026-05-02", 250); err != nil {
		t.Fatalf("Contribute(Achieng, 2026-05-02, 250) returned an error: %v", err)
	}
	want := []Contribution{{Date: "2026-04-02", Amount: 1000}, {Date: "2026-05-02", Amount: 250}}
	if got := l.Members[1].Contributions; !reflect.DeepEqual(got, want) {
		t.Errorf("Achieng's contributions\n got: %+v\nwant: %+v", got, want)
	}

	bad := []struct {
		name, date string
		amount     int
	}{
		{"Kamau", "2026-04-02", 500},      // not a member
		{"Wanjiku", "2026-04-02", 0},      // nothing paid
		{"Wanjiku", "02/04/2026", 500},    // wrong date format
		{"Wanjiku", "2026-02-30", 500},    // no such day
		{"Wanjiku", "2026-4-2", 500},      // missing zeros
		{"Wanjiku", "2026-04-02x", 500},   // trailing junk
		{"wanjiku", "2026-04-02", 500},    // names are exact
		{"Wanjiku", "", 500},              // no date
		{"Wanjiku", "2026-04-02", -100},   // negative
		{"Achieng ", "2026-04-02", 500},   // trailing space
		{"Wanjiku", "2026-13-01", 500},    // no such month
		{"Wanjiku", "2026-00-10", 500},    // no such month
		{"Wanjiku", "2026-04-31", 500},    // April has 30 days
		{"Wanjiku", "26-04-02", 500},      // two-digit year
		{"Wanjiku", "2026/04/02", 500},    // slashes
		{"Wanjiku", " 2026-04-02", 500},   // leading space
		{"Wanjiku", "2026-04-02 ", 500},   // trailing space
		{"Wanjiku", "2026-04-02T10", 500}, // a time too
	}
	for _, b := range bad {
		l := sample()
		before := l.Total(b.name)
		if err := l.Contribute(b.name, b.date, b.amount); err == nil {
			t.Errorf("Contribute(%q, %q, %d) = nil, want an error", b.name, b.date, b.amount)
		}
		if after := l.Total(b.name); after != before {
			t.Errorf("a failed Contribute(%q, %q, %d) changed the total from %d to %d", b.name, b.date, b.amount, before, after)
		}
	}
}

func TestTotal(t *testing.T) {
	l := sample()
	tests := []struct {
		name string
		want int
	}{
		{"Wanjiku", 1250},
		{"Achieng", 0},
		{"Kamau", 0},
	}
	for _, tt := range tests {
		if got := l.Total(tt.name); got != tt.want {
			t.Errorf("Total(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestSaveWritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := Save(path, sample()); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the saved file: %v", err)
	}
	if string(got) != sampleJSON {
		t.Errorf("the saved file\n got:\n%s\nwant:\n%s", got, sampleJSON)
	}
	if strings.Contains(string(got), "4321") || strings.Contains(strings.ToLower(string(got)), "pin") {
		t.Errorf("the saved file contains a PIN; tag the field `json:\"-\"`")
	}
}

func TestSaveEmptyLedger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := Save(path, Ledger{Chama: "New Chama"}); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}
	got, _ := os.ReadFile(path)
	want := "{\n  \"chama\": \"New Chama\",\n  \"members\": []\n}\n"
	if string(got) != want {
		t.Errorf("an empty ledger\n got: %q\nwant: %q", got, want)
	}
}

func TestSaveDoesNotChangeTheCaller(t *testing.T) {
	l := Ledger{Chama: "Umoja", Members: []Member{{Name: "Achieng"}}}
	if err := Save(filepath.Join(t.TempDir(), "ledger.json"), l); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}
	if l.Members[0].Contributions != nil {
		t.Errorf("Save changed the caller's ledger: Contributions is now %#v", l.Members[0].Contributions)
	}
}

func TestSaveLeavesOneFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.json")
	for i := 0; i < 3; i++ { // saving again replaces the file
		if err := Save(path, sample()); err != nil {
			t.Fatalf("Save #%d returned an error: %v", i+1, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if !reflect.DeepEqual(names, []string{"ledger.json"}) {
		t.Errorf("after three Saves the folder holds %q, want only ledger.json (remove temporary files)", names)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("ledger.json has permissions %v, want -rw------- (0600)", perm)
		}
	}
}

func TestSaveToMissingFolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-folder", "ledger.json")
	if err := Save(path, sample()); err == nil {
		t.Errorf("Save into a folder that doesn't exist = nil, want an error")
	}
}

func TestLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := Save(path, sample()); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	want := sample()
	want.Members[0].PIN = "" // never saved, so it can't come back
	want.Members[1].PIN = ""
	want.Members[1].Contributions = []Contribution{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load(Save(sample()))\n got: %+v\nwant: %+v", got, want)
	}
}

func writeFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{"not JSON", "chama: Umoja"},
		{"cut off halfway", `{"chama": "Umoja", "members": [`},
		{"a typo in a field name", `{"chama": "Umoja", "members": [{"name": "Achieng", "contributions": [{"date": "2026-04-01", "ammount_ksh": 500}]}]}`},
		{"wrong type", `{"chama": "Umoja", "members": [{"name": "Achieng", "contributions": [{"date": "2026-04-01", "amount_ksh": "500"}]}]}`},
		{"empty file", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFile(t, tt.contents)
			_, err := Load(path)
			if err == nil {
				t.Fatalf("Load(%q) = nil error, want an error", tt.contents)
			}
			if !strings.Contains(err.Error(), path) {
				t.Errorf("Load error %q doesn't say which file", err)
			}
		})
	}
}

func TestLoadMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := Load(path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load of a missing file = %v, want an error wrapping fs.ErrNotExist", err)
	}
}

func TestLoadOrNew(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	got, err := LoadOrNew(path, "Umoja")
	if err != nil {
		t.Fatalf("LoadOrNew of a missing file returned an error: %v", err)
	}
	if got.Chama != "Umoja" || got.Members == nil || len(got.Members) != 0 {
		t.Errorf("LoadOrNew of a missing file = %#v, want Ledger{Chama: \"Umoja\", Members: []Member{}}", got)
	}

	if err := Save(path, sample()); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}
	got, err = LoadOrNew(path, "Ignored")
	if err != nil || got.Chama != "Umoja Women's Group" || len(got.Members) != 2 {
		t.Errorf("LoadOrNew of a saved file = %+v, %v; want the saved ledger", got, err)
	}

	bad := writeFile(t, "{")
	if _, err := LoadOrNew(bad, "Umoja"); err == nil {
		t.Errorf("LoadOrNew of a broken file = nil error; only a missing file means start fresh")
	}
}
