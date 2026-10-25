// Package saveload keeps a chama's contribution ledger in a JSON file, so it survives
// after the program exits.
package saveload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Contribution is one payment into the chama.
type Contribution struct {
	Date   string `json:"date"`       // YYYY-MM-DD
	Amount int    `json:"amount_ksh"` // whole shillings
}

// Member is one member of the chama. PIN is only for this program's memory: it is
// never written to the file.
type Member struct {
	Name          string         `json:"name"`
	Phone         string         `json:"phone"`
	Contributions []Contribution `json:"contributions"`
	Notes         string         `json:"notes,omitempty"`
	PIN           string         `json:"-"`
}

// Ledger is everything the file holds.
type Ledger struct {
	Chama   string   `json:"chama"`
	Members []Member `json:"members"`
}

// Contribute records that the member called name paid amount shillings on date. It
// returns an error if there's no such member, amount is below 1, or date isn't a real
// date in the form YYYY-MM-DD.
func (l *Ledger) Contribute(name, date string, amount int) error {
	if amount < 1 {
		return fmt.Errorf("contribution of Ksh%d: must be at least Ksh1", amount)
	}
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return fmt.Errorf("date %q: want YYYY-MM-DD", date)
	}
	for i := range l.Members {
		if l.Members[i].Name == name {
			m := &l.Members[i]
			m.Contributions = append(m.Contributions, Contribution{Date: date, Amount: amount})
			return nil
		}
	}
	return fmt.Errorf("no member called %q", name)
}

// Total returns how much the member called name has paid in, or 0 if there's no such
// member.
func (l Ledger) Total(name string) int {
	total := 0
	for _, m := range l.Members {
		if m.Name == name {
			for _, c := range m.Contributions {
				total += c.Amount
			}
		}
	}
	return total
}

// Save writes l to path as JSON indented with two spaces, ending with a newline.
// Empty member and contribution lists are written as [] rather than null.
//
// It writes to a temporary file in the same folder first and then renames it over
// path, so a crash halfway through never leaves a half-written ledger. The file is
// readable and writable by its owner only (0600), because it holds phone numbers.
func Save(path string, l Ledger) error {
	if l.Members == nil {
		l.Members = []Member{}
	}
	members := make([]Member, len(l.Members))
	for i, m := range l.Members {
		if m.Contributions == nil {
			m.Contributions = []Contribution{}
		}
		members[i] = m
	}
	l.Members = members

	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	defer os.Remove(tmp.Name()) // does nothing once the rename has happened

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("save %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}

// Load reads a ledger that Save wrote. It returns an error if the file can't be read
// (a missing file's error wraps fs.ErrNotExist, so callers can check with errors.Is),
// isn't valid JSON, or has a field the Ledger doesn't know, such as a typo like
// "ammount_ksh". Every error starts with the path.
func Load(path string) (Ledger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Ledger{}, fmt.Errorf("load: %w", err)
	}
	var l Ledger
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&l); err != nil {
		return Ledger{}, fmt.Errorf("load %s: %w", path, err)
	}
	return l, nil
}

// LoadOrNew loads the ledger at path. If there's no file yet, it returns a new, empty
// ledger for the chama called chama, with Members set to an empty slice. Any other
// error is returned.
func LoadOrNew(path, chama string) (Ledger, error) {
	l, err := Load(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Ledger{Chama: chama, Members: []Member{}}, nil
	}
	return l, err
}
