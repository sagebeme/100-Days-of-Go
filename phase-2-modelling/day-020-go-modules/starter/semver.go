// Package gomodules reads and compares semantic versions the way the go command
// does: "v1.4.2", "v2.0.0-rc.1". It's the logic behind go get choosing an upgrade.
package gomodules

// Version is a parsed semantic version.
type Version struct {
	Major, Minor, Patch int
	Pre                 string // the pre-release part after "-", e.g. "rc.1"; "" for a release
}

// Parse reads a version in the form Go modules use: a "v", three numbers separated by
// dots, and optionally "-" and a pre-release such as "rc.1" or "beta".
//
// It returns an error if the "v" is missing, a number is missing, has a leading zero
// ("v1.02.0") or isn't a number, or the pre-release is malformed. A pre-release is one
// or more dot-separated identifiers made of letters, digits and "-"; an identifier
// can't be empty, and one made only of digits can't have a leading zero. Build
// metadata ("+build.5") is not supported and is an error too.
func Parse(s string) (Version, error) {
	// Your code here.
	return Version{}, nil
}

// String writes the version back out: "v1.2.3" or "v1.2.3-rc.1".
func (v Version) String() string {
	// Your code here.
	return ""
}

// Compare returns -1 if a comes before b, 0 if they're equal and +1 if a comes after.
// It follows semver precedence:
//
//   - compare Major, then Minor, then Patch, as numbers
//   - if those are equal, a pre-release comes before the release: v1.0.0-rc.1 < v1.0.0
//   - two pre-releases are compared identifier by identifier: numeric identifiers as
//     numbers, others as strings, and a numeric one before a non-numeric one; if one
//     runs out first, it comes first (v1.0.0-alpha < v1.0.0-alpha.1)
func Compare(a, b Version) int {
	// Your code here.
	return 0
}

// Upgrade picks what `go get -u` would move current to from the available versions:
// the highest one with the same major version that is newer than current and is not
// a pre-release. If there's nothing newer it returns current unchanged. A new major
// version is never picked: in Go, v2 and up are a different module path.
//
// It returns an error if current or any available version doesn't parse.
func Upgrade(current string, available []string) (string, error) {
	// Your code here.
	return "", nil
}

// ModulePath returns the import path for a module's major version. Versions 0 and 1
// use the plain path; from v2 on, Go adds "/vN" to the end:
//
//	ModulePath("github.com/sagebeme/pesa", 1) → "github.com/sagebeme/pesa"
//	ModulePath("github.com/sagebeme/pesa", 3) → "github.com/sagebeme/pesa/v3"
func ModulePath(base string, major int) string {
	// Your code here.
	return ""
}
