// Package course finds the day folders in the repo.
package course

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var dayDir = regexp.MustCompile(`^day-(\d{3})-`)

// Day is one written day's folder.
type Day struct {
	N    int
	Dir  string // relative to the repo root, with forward slashes
	Name string // the folder name, e.g. day-001-mpesa-sms
}

// Root walks up from the working directory to the folder holding go.mod.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("can't find go.mod: run this from inside the course repo")
		}
		dir = parent
	}
}

// Days lists every phase-*/day-NNN-* folder, in day order.
func Days(root string) ([]Day, error) {
	phases, err := filepath.Glob(filepath.Join(root, "phase-*"))
	if err != nil {
		return nil, err
	}
	var days []Day
	for _, p := range phases {
		entries, err := os.ReadDir(p)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			m := dayDir.FindStringSubmatch(e.Name())
			if !e.IsDir() || m == nil {
				continue
			}
			n, _ := strconv.Atoi(m[1])
			rel, _ := filepath.Rel(root, filepath.Join(p, e.Name()))
			days = append(days, Day{N: n, Dir: filepath.ToSlash(rel), Name: e.Name()})
		}
	}
	sort.Slice(days, func(i, j int) bool { return days[i].N < days[j].N })
	return days, nil
}

// Find returns the day numbered n.
func Find(root string, n int) (Day, error) {
	days, err := Days(root)
	if err != nil {
		return Day{}, err
	}
	for _, d := range days {
		if d.N == n {
			return d, nil
		}
	}
	return Day{}, fmt.Errorf("there's no folder for day %d yet", n)
}

// Exists reports whether path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
