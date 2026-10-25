// Package migrations changes the booking API's database schema one numbered step at
// a time. Each step is a pair of SQL files, one to apply it and one to undo it. A
// schema_migrations table records which steps the database has, and each step runs
// inside a transaction, so it either happens completely or not at all.
package migrations

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Migration is one versioned schema change.
type Migration struct {
	Version int
	Name    string
	Up      string // SQL that applies the change
	Down    string // SQL that undoes it
}

// Parse reads the migrations in dir of fsys. Each file is named
// NNNN_name.up.sql or NNNN_name.down.sql: a version of one or more digits above 0,
// an underscore, a name of lower-case letters, digits and underscores, then the
// direction. Files that don't end in .sql are ignored, so a README can live there.
// It returns the migrations sorted by version, or an error that names the file when:
//
//   - a .sql file doesn't follow the pattern
//   - a version has two up files or two down files
//   - a version's up and down files have different names
//   - a version has no up file, or no down file
//   - an up file is empty
func Parse(fsys fs.FS, dir string) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	byVersion := map[int]*Migration{}
	hasUp := map[int]bool{}
	hasDown := map[int]bool{}
	for _, e := range entries {
		file := e.Name()
		if e.IsDir() || !strings.HasSuffix(file, ".sql") {
			continue
		}
		version, name, up, err := parseName(file)
		if err != nil {
			return nil, err
		}
		body, err := fs.ReadFile(fsys, path.Join(dir, file))
		if err != nil {
			return nil, err
		}

		m, ok := byVersion[version]
		if !ok {
			m = &Migration{Version: version, Name: name}
			byVersion[version] = m
		}
		if m.Name != name {
			return nil, fmt.Errorf("%s: version %d is already called %q", file, version, m.Name)
		}
		if up {
			if hasUp[version] {
				return nil, fmt.Errorf("%s: version %d has two up files", file, version)
			}
			if strings.TrimSpace(string(body)) == "" {
				return nil, fmt.Errorf("%s: the up file is empty", file)
			}
			hasUp[version] = true
			m.Up = string(body)
		} else {
			if hasDown[version] {
				return nil, fmt.Errorf("%s: version %d has two down files", file, version)
			}
			hasDown[version] = true
			m.Down = string(body)
		}
	}

	out := make([]Migration, 0, len(byVersion))
	for v, m := range byVersion {
		if !hasUp[v] {
			return nil, fmt.Errorf("version %d (%s) has no up file", v, m.Name)
		}
		if !hasDown[v] {
			return nil, fmt.Errorf("version %d (%s) has no down file", v, m.Name)
		}
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// parseName splits "0002_create_trips.up.sql" into 2, "create_trips", true.
func parseName(file string) (version int, name string, up bool, err error) {
	bad := fmt.Errorf("%s: want a name like 0001_create_routes.up.sql or 0001_create_routes.down.sql", file)
	var rest string
	switch {
	case strings.HasSuffix(file, ".up.sql"):
		rest, up = strings.TrimSuffix(file, ".up.sql"), true
	case strings.HasSuffix(file, ".down.sql"):
		rest = strings.TrimSuffix(file, ".down.sql")
	default:
		return 0, "", false, bad
	}
	digits, name, ok := strings.Cut(rest, "_")
	if !ok || digits == "" || name == "" {
		return 0, "", false, bad
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, "", false, bad
		}
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return 0, "", false, bad
		}
	}
	version, err = strconv.Atoi(digits)
	if err != nil || version < 1 {
		return 0, "", false, bad
	}
	return version, name, up, nil
}

// ErrUnknownVersion means the database has a version no migration file knows about:
// someone applied a migration from a branch, or a file was deleted.
var ErrUnknownVersion = errors.New("the database has a version with no migration file")

// Pending returns the migrations in all that aren't in applied, in version order.
// all must be sorted, as Parse returns it. It returns an error wrapping
// ErrUnknownVersion if applied holds a version that isn't in all, and an error if a
// pending migration is older than the newest applied one: that file was added after
// a later migration had already run, so applying it now could do things in the
// wrong order.
func Pending(all []Migration, applied []int) ([]Migration, error) {
	known := map[int]bool{}
	for _, m := range all {
		known[m.Version] = true
	}
	done := map[int]bool{}
	newest := 0
	for _, v := range applied {
		if !known[v] {
			return nil, fmt.Errorf("version %d: %w", v, ErrUnknownVersion)
		}
		done[v] = true
		newest = max(newest, v)
	}

	pending := []Migration{}
	for _, m := range all {
		if done[m.Version] {
			continue
		}
		if m.Version < newest {
			return nil, fmt.Errorf("version %d (%s) is older than version %d, which is already applied", m.Version, m.Name, newest)
		}
		pending = append(pending, m)
	}
	return pending, nil
}

// ToRollBack returns the migrations to undo to go back steps versions: the newest
// applied ones first. If steps is more than len(applied), it returns them all. It
// returns an error if steps is below 1, and an error wrapping ErrUnknownVersion if
// applied holds a version that isn't in all.
func ToRollBack(all []Migration, applied []int, steps int) ([]Migration, error) {
	if steps < 1 {
		return nil, fmt.Errorf("steps must be 1 or more, got %d", steps)
	}
	byVersion := map[int]Migration{}
	for _, m := range all {
		byVersion[m.Version] = m
	}
	versions := append([]int(nil), applied...)
	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	out := []Migration{}
	for _, v := range versions {
		if len(out) == steps {
			break
		}
		m, ok := byVersion[v]
		if !ok {
			return nil, fmt.Errorf("version %d: %w", v, ErrUnknownVersion)
		}
		out = append(out, m)
	}
	return out, nil
}
