// Package keyvaluestore is a tiny key-value database. Every change is appended to a
// write-ahead log on disk before it reaches the in-memory map, so the data survives a
// restart. Open replays the log, drops a torn or corrupt last record, and Compact
// rewrites the log so it holds only the live keys.
package keyvaluestore

import "errors"

// LogFile is the name of the write-ahead log inside the store's folder.
const LogFile = "data.wal"

// MaxKeySize and MaxValueSize bound one record, so a corrupt length can't make Open
// allocate gigabytes.
const (
	MaxKeySize   = 1 << 10
	MaxValueSize = 1 << 20
)

var (
	// ErrClosed is returned by every method called after Close.
	ErrClosed = errors.New("keyvaluestore: store is closed")
	// ErrEmptyKey is returned by Set and Delete for the key "".
	ErrEmptyKey = errors.New("keyvaluestore: empty key")
	// ErrTooLarge is returned by Set when the key or value is over its limit.
	ErrTooLarge = errors.New("keyvaluestore: key or value too large")
)

// Store is a key-value store backed by one append-only log file. It is safe for
// concurrent use. Add the fields you need: a mutex, the open file, the map, the size.
type Store struct {
}

// Open opens the store in dir, creating the folder and the log if needed, and replays
// the log. A torn or corrupt record at the end (a crash mid-write) is cut off, so the
// next write starts on a clean boundary.
func Open(dir string) (*Store, error) {
	// Your code here
	return &Store{}, nil
}

// Get returns the value for key and whether it was there.
func (s *Store) Get(key string) (string, bool) {
	// Your code here
	return "", false
}

// Set stores value under key. The record is on disk (synced) before Set returns.
// It returns ErrEmptyKey, ErrTooLarge or ErrClosed when those apply.
func (s *Store) Set(key, value string) error {
	// Your code here
	return nil
}

// Delete removes key. Deleting a key that isn't there is not an error.
func (s *Store) Delete(key string) error {
	// Your code here
	return nil
}

// Len returns the number of live keys.
func (s *Store) Len() int {
	// Your code here
	return 0
}

// Keys returns the live keys in sorted order.
func (s *Store) Keys() []string {
	// Your code here
	return nil
}

// Size returns the size of the log file in bytes.
func (s *Store) Size() int64 {
	// Your code here
	return 0
}

// Compact rewrites the log so it holds one record per live key. Write a new file,
// sync it, and rename it over the old log, so a crash at any point leaves either the
// old log or the new one.
func (s *Store) Compact() error {
	// Your code here
	return nil
}

// Close closes the log. Calling Close twice returns ErrClosed.
func (s *Store) Close() error {
	// Your code here
	return nil
}
