package keyvaluestore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

// open opens a store in dir and closes it when the test ends.
func open(t *testing.T, dir string) *Store {
	t.Helper()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%q): %v", dir, err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// mustSet fails the test if Set returns an error.
func mustSet(t *testing.T, s *Store, key, value string) {
	t.Helper()
	if err := s.Set(key, value); err != nil {
		t.Fatalf("Set(%q, %q): %v", key, value, err)
	}
}

// wantGet checks one key. want "" with ok false means the key must be missing.
func wantGet(t *testing.T, s *Store, key, want string, wantOK bool) {
	t.Helper()
	got, ok := s.Get(key)
	if got != want || ok != wantOK {
		t.Errorf("Get(%q) = %q, %v; want %q, %v", key, got, ok, want, wantOK)
	}
}

func logPath(dir string) string { return filepath.Join(dir, LogFile) }

// fileSize returns the size of the file at path, failing the test if it's missing.
func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("the log should exist after a Set: %v", err)
	}
	return info.Size()
}

func TestSetGetDelete(t *testing.T) {
	s := open(t, t.TempDir())
	mustSet(t, s, "wanjiku", "0712345678")
	mustSet(t, s, "otieno", "0722000111")
	mustSet(t, s, "wanjiku", "0799999999") // overwrite
	if err := s.Delete("otieno"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete("nobody"); err != nil {
		t.Errorf("Delete of a missing key = %v, want nil", err)
	}

	tests := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"wanjiku", "0799999999", true},
		{"otieno", "", false},
		{"nobody", "", false},
	}
	for _, tt := range tests {
		wantGet(t, s, tt.key, tt.want, tt.wantOK)
	}
	if got := s.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestEmptyValueIsAValue(t *testing.T) {
	s := open(t, t.TempDir())
	mustSet(t, s, "note", "")
	wantGet(t, s, "note", "", true)
}

func TestBadInput(t *testing.T) {
	s := open(t, t.TempDir())
	if err := s.Set("", "x"); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Set(\"\", ...) = %v, want ErrEmptyKey", err)
	}
	if err := s.Delete(""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Delete(\"\") = %v, want ErrEmptyKey", err)
	}
	big := string(make([]byte, MaxValueSize+1))
	if err := s.Set("k", big); !errors.Is(err, ErrTooLarge) {
		t.Errorf("Set with a value over MaxValueSize = %v, want ErrTooLarge", err)
	}
}

func TestSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	mustSet(t, s, "chama:balance", "Ksh 45,000")
	mustSet(t, s, "chama:members", "12")
	mustSet(t, s, "chama:members", "13")
	mustSet(t, s, "temp", "x")
	if err := s.Delete("temp"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2 := open(t, dir)
	wantGet(t, s2, "chama:balance", "Ksh 45,000", true)
	wantGet(t, s2, "chama:members", "13", true)
	wantGet(t, s2, "temp", "", false)
	if got, want := s2.Keys(), []string{"chama:balance", "chama:members"}; !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestSizeMatchesFile(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir)
	for i := range 10 {
		mustSet(t, s, fmt.Sprintf("k%d", i), "value")
	}
	if size := fileSize(t, logPath(dir)); s.Size() == 0 || s.Size() != size {
		t.Errorf("Size() = %d, file is %d bytes; want them equal and non-zero", s.Size(), size)
	}
}

// crash writes three keys, closes the store, and lets damage change the log file the
// way a crash mid-write would.
func crash(t *testing.T, damage func(t *testing.T, path string)) string {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	mustSet(t, s, "a", "alpha")
	mustSet(t, s, "b", "bravo")
	mustSet(t, s, "c", "charlie")
	s.Close()
	damage(t, logPath(dir))
	return dir
}

func TestCrashRecovery(t *testing.T) {
	tests := []struct {
		name   string
		damage func(t *testing.T, path string)
	}{
		{"last record cut short by one byte", func(t *testing.T, path string) {
			if err := os.Truncate(path, fileSize(t, path)-1); err != nil {
				t.Fatal(err)
			}
		}},
		{"last record cut to half", func(t *testing.T, path string) {
			whole := fileSize(t, path) / 3 // three records of about the same size
			if err := os.Truncate(path, 2*whole+whole/2); err != nil {
				t.Fatal(err)
			}
		}},
		{"last byte flipped", func(t *testing.T, path string) {
			b, _ := os.ReadFile(path)
			if len(b) == 0 {
				t.Fatal("the log is empty after three Sets")
			}
			b[len(b)-1] ^= 0xFF
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"two stray bytes after the last record", func(t *testing.T, path string) {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				t.Fatalf("the log should exist after a Set: %v", err)
			}
			f.Write([]byte{0x01, 0x02})
			f.Close()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := crash(t, tt.damage)
			s, err := Open(dir)
			if err != nil {
				t.Fatalf("Open after a crash: %v, want the good records back", err)
			}
			wantGet(t, s, "a", "alpha", true)
			wantGet(t, s, "b", "bravo", true)
			if v, ok := s.Get("c"); ok && v != "charlie" {
				t.Errorf("Get(\"c\") = %q after damage; a damaged record must never return a wrong value", v)
			}

			// The next write must land on a clean boundary, so it survives another restart.
			mustSet(t, s, "d", "delta")
			s.Close()
			s2 := open(t, dir)
			wantGet(t, s2, "a", "alpha", true)
			wantGet(t, s2, "d", "delta", true)
		})
	}
}

func TestDamagedRecordIsDropped(t *testing.T) {
	dir := crash(t, func(t *testing.T, path string) {
		os.Truncate(path, fileSize(t, path)-1)
	})
	s := open(t, dir)
	wantGet(t, s, "c", "", false)
	if got := s.Len(); got != 2 {
		t.Errorf("Len() = %d after losing the last record, want 2", got)
	}
}

func TestCompact(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 50 {
		mustSet(t, s, "float", fmt.Sprintf("Ksh %d", i*100))
	}
	mustSet(t, s, "gone", "soon")
	if err := s.Delete("gone"); err != nil {
		t.Fatal(err)
	}
	mustSet(t, s, "till", "247247")
	before := s.Size()

	if err := s.Compact(); err != nil {
		t.Fatalf("Compact: %v", err)
	}
	after := s.Size()
	if after <= 0 || after*10 > before {
		t.Errorf("Size() went from %d to %d bytes; want it at least 10 times smaller", before, after)
	}
	wantGet(t, s, "float", "Ksh 4900", true)
	wantGet(t, s, "gone", "", false)

	mustSet(t, s, "after", "compact") // the store must still write to the new log
	s.Close()

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != LogFile {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("folder holds %v after Compact and Close, want only %s", names, LogFile)
	}

	s2 := open(t, dir)
	wantGet(t, s2, "float", "Ksh 4900", true)
	wantGet(t, s2, "till", "247247", true)
	wantGet(t, s2, "after", "compact", true)
	wantGet(t, s2, "gone", "", false)
}

func TestClosed(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mustSet(t, s, "k", "v")
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := s.Set("k", "v2"); !errors.Is(err, ErrClosed) {
		t.Errorf("Set after Close = %v, want ErrClosed", err)
	}
	if err := s.Delete("k"); !errors.Is(err, ErrClosed) {
		t.Errorf("Delete after Close = %v, want ErrClosed", err)
	}
	if err := s.Compact(); !errors.Is(err, ErrClosed) {
		t.Errorf("Compact after Close = %v, want ErrClosed", err)
	}
	if err := s.Close(); !errors.Is(err, ErrClosed) {
		t.Errorf("second Close = %v, want ErrClosed", err)
	}
}

func TestConcurrentWriters(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for w := range 8 {
		wg.Go(func() {
			for i := range 10 {
				if err := s.Set(fmt.Sprintf("w%d-%d", w, i), "x"); err != nil {
					t.Error(err)
				}
				s.Get("w0-0")
			}
		})
	}
	wg.Wait()
	if got := s.Len(); got != 80 {
		t.Errorf("Len() = %d, want 80", got)
	}
	s.Close()
	s2 := open(t, dir)
	if got := s2.Len(); got != 80 {
		t.Errorf("Len() after restart = %d, want 80", got)
	}
}
