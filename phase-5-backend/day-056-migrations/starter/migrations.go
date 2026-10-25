// Package migrations changes the booking API's database schema one numbered step at
// a time. Each step is a pair of SQL files, one to apply it and one to undo it. A
// schema_migrations table records which steps the database has, and each step runs
// inside a transaction, so it either happens completely or not at all.
package migrations

import (
	"errors"
	"io/fs"
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
	// Your code here. Follow the steps in the README.
	return nil, nil
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
	// Your code here.
	return nil, nil
}

// ToRollBack returns the migrations to undo to go back steps versions: the newest
// applied ones first. If steps is more than len(applied), it returns them all. It
// returns an error if steps is below 1, and an error wrapping ErrUnknownVersion if
// applied holds a version that isn't in all.
func ToRollBack(all []Migration, applied []int, steps int) ([]Migration, error) {
	// Your code here.
	return nil, nil
}
