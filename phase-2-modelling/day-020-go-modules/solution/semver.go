// Package gomodules reads and compares semantic versions the way the go command
// does: "v1.4.2", "v2.0.0-rc.1". It's the logic behind go get choosing an upgrade.
package gomodules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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
	rest, ok := strings.CutPrefix(s, "v")
	if !ok {
		return Version{}, fmt.Errorf("version %q must start with v", s)
	}
	if strings.Contains(rest, "+") {
		return Version{}, fmt.Errorf("version %q: build metadata isn't supported", s)
	}
	core, pre, hasPre := strings.Cut(rest, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("version %q needs three numbers, like v1.2.3", s)
	}
	var nums [3]int
	for i, p := range parts {
		n, err := number(p)
		if err != nil {
			return Version{}, fmt.Errorf("version %q: %w", s, err)
		}
		nums[i] = n
	}
	if hasPre {
		if err := checkPre(pre); err != nil {
			return Version{}, fmt.Errorf("version %q: %w", s, err)
		}
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Pre: pre}, nil
}

// number reads a non-negative whole number with no sign and no leading zero.
func number(s string) (int, error) {
	if s == "" {
		return 0, errors.New("missing number")
	}
	if !allDigits(s) {
		return 0, fmt.Errorf("%q isn't a number", s)
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("%q has a leading zero", s)
	}
	return strconv.Atoi(s)
}

func checkPre(pre string) error {
	for _, id := range strings.Split(pre, ".") {
		if id == "" {
			return errors.New("empty pre-release identifier")
		}
		for i := 0; i < len(id); i++ {
			c := id[i]
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return fmt.Errorf("pre-release %q has a bad character %q", pre, c)
			}
		}
		if allDigits(id) && len(id) > 1 && id[0] == '0' {
			return fmt.Errorf("pre-release identifier %q has a leading zero", id)
		}
	}
	return nil
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// String writes the version back out: "v1.2.3" or "v1.2.3-rc.1".
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
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
	if c := compareInt(a.Major, b.Major); c != 0 {
		return c
	}
	if c := compareInt(a.Minor, b.Minor); c != 0 {
		return c
	}
	if c := compareInt(a.Patch, b.Patch); c != 0 {
		return c
	}
	switch {
	case a.Pre == b.Pre:
		return 0
	case a.Pre == "":
		return 1
	case b.Pre == "":
		return -1
	}
	as, bs := strings.Split(a.Pre, "."), strings.Split(b.Pre, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if c := compareIdent(as[i], bs[i]); c != 0 {
			return c
		}
	}
	return compareInt(len(as), len(bs))
}

func compareIdent(a, b string) int {
	an, bn := allDigits(a), allDigits(b)
	switch {
	case an && bn:
		// Same digits count first, so "10" > "9" without converting.
		if c := compareInt(len(a), len(b)); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	case an:
		return -1
	case bn:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// Upgrade picks what `go get -u` would move current to from the available versions:
// the highest one with the same major version that is newer than current and is not
// a pre-release. If there's nothing newer it returns current unchanged. A new major
// version is never picked: in Go, v2 and up are a different module path.
//
// It returns an error if current or any available version doesn't parse.
func Upgrade(current string, available []string) (string, error) {
	cur, err := Parse(current)
	if err != nil {
		return "", err
	}
	best := cur
	for _, s := range available {
		v, err := Parse(s)
		if err != nil {
			return "", err
		}
		if v.Major == cur.Major && v.Pre == "" && Compare(v, best) > 0 {
			best = v
		}
	}
	return best.String(), nil
}

// ModulePath returns the import path for a module's major version. Versions 0 and 1
// use the plain path; from v2 on, Go adds "/vN" to the end:
//
//	ModulePath("github.com/sagebeme/pesa", 1) → "github.com/sagebeme/pesa"
//	ModulePath("github.com/sagebeme/pesa", 3) → "github.com/sagebeme/pesa/v3"
func ModulePath(base string, major int) string {
	if major < 2 {
		return base
	}
	return fmt.Sprintf("%s/v%d", base, major)
}
