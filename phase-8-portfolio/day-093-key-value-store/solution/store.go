// Package keyvaluestore is a tiny key-value database. Every change is appended to a
// write-ahead log on disk before it reaches the in-memory map, so the data survives a
// restart. Open replays the log, drops a torn or corrupt last record, and Compact
// rewrites the log so it holds only the live keys.
package keyvaluestore

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

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

// Record layout, all integers big-endian:
//
//	crc32 (4) | op (1) | key length (4) | value length (4) | key | value
//
// The CRC-32 (IEEE) covers everything after itself.
const headerSize = 4 + 1 + 4 + 4

const (
	opSet byte = 1
	opDel byte = 2
)

// Store is a key-value store backed by one append-only log file. It is safe for
// concurrent use.
type Store struct {
	mu     sync.RWMutex
	dir    string
	f      *os.File
	data   map[string]string
	size   int64
	closed bool
}

// Open opens the store in dir, creating the folder and the log if needed, and replays
// the log. A torn or corrupt record at the end (a crash mid-write) is cut off, so the
// next write starts on a clean boundary.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, LogFile), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	data, good, err := replay(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Truncate(good); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := f.Seek(good, io.SeekStart); err != nil {
		f.Close()
		return nil, err
	}
	return &Store{dir: dir, f: f, data: data, size: good}, nil
}

// replay reads records until the end or the first bad one. It returns the map and the
// offset just after the last good record.
func replay(r io.Reader) (map[string]string, int64, error) {
	data := map[string]string{}
	br := bufio.NewReader(r)
	var good int64
	for {
		op, key, value, n, err := readRecord(br)
		if err == io.EOF || errors.Is(err, errCorrupt) {
			return data, good, nil
		}
		if err != nil {
			return nil, 0, err
		}
		switch op {
		case opSet:
			data[key] = value
		case opDel:
			delete(data, key)
		}
		good += n
	}
}

var errCorrupt = errors.New("corrupt record")

// readRecord reads one record. A short read or a bad checksum is errCorrupt; a clean
// end of file before any byte is io.EOF.
func readRecord(br *bufio.Reader) (op byte, key, value string, n int64, err error) {
	var h [headerSize]byte
	if _, err := io.ReadFull(br, h[:]); err != nil {
		if err == io.EOF {
			return 0, "", "", 0, io.EOF
		}
		if err == io.ErrUnexpectedEOF {
			return 0, "", "", 0, errCorrupt
		}
		return 0, "", "", 0, err
	}
	sum := binary.BigEndian.Uint32(h[0:4])
	op = h[4]
	klen := binary.BigEndian.Uint32(h[5:9])
	vlen := binary.BigEndian.Uint32(h[9:13])
	if (op != opSet && op != opDel) || klen == 0 || klen > MaxKeySize || vlen > MaxValueSize {
		return 0, "", "", 0, errCorrupt
	}
	body := make([]byte, klen+vlen)
	if _, err := io.ReadFull(br, body); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return 0, "", "", 0, errCorrupt
		}
		return 0, "", "", 0, err
	}
	crc := crc32.NewIEEE()
	crc.Write(h[4:])
	crc.Write(body)
	if crc.Sum32() != sum {
		return 0, "", "", 0, errCorrupt
	}
	return op, string(body[:klen]), string(body[klen:]), int64(headerSize) + int64(len(body)), nil
}

// encode builds one record.
func encode(op byte, key, value string) []byte {
	b := make([]byte, headerSize, headerSize+len(key)+len(value))
	b[4] = op
	binary.BigEndian.PutUint32(b[5:9], uint32(len(key)))
	binary.BigEndian.PutUint32(b[9:13], uint32(len(value)))
	b = append(b, key...)
	b = append(b, value...)
	binary.BigEndian.PutUint32(b[0:4], crc32.ChecksumIEEE(b[4:]))
	return b
}

// write appends a record and syncs it to disk. The caller holds the write lock.
func (s *Store) write(rec []byte) error {
	n, err := s.f.Write(rec)
	s.size += int64(n)
	if err != nil {
		return err
	}
	return s.f.Sync()
}

// Get returns the value for key and whether it was there.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return "", false
	}
	v, ok := s.data[key]
	return v, ok
}

// Set stores value under key. The record is on disk before Set returns.
func (s *Store) Set(key, value string) error {
	if key == "" {
		return ErrEmptyKey
	}
	if len(key) > MaxKeySize || len(value) > MaxValueSize {
		return ErrTooLarge
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	if err := s.write(encode(opSet, key, value)); err != nil {
		return fmt.Errorf("keyvaluestore: set %q: %w", key, err)
	}
	s.data[key] = value
	return nil
}

// Delete removes key. Deleting a key that isn't there is not an error and writes nothing.
func (s *Store) Delete(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	if _, ok := s.data[key]; !ok {
		return nil
	}
	if err := s.write(encode(opDel, key, "")); err != nil {
		return fmt.Errorf("keyvaluestore: delete %q: %w", key, err)
	}
	delete(s.data, key)
	return nil
}

// Len returns the number of live keys.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Keys returns the live keys in sorted order.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Size returns the size of the log file in bytes.
func (s *Store) Size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

// Compact rewrites the log so it holds one set record per live key. It writes a new
// file, syncs it, and renames it over the old log, so a crash at any point leaves
// either the old log or the new one, never half of each.
func (s *Store) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	tmpPath := filepath.Join(s.dir, LogFile+".tmp")
	tmp, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	cleanup := func(err error) error {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("keyvaluestore: compact: %w", err)
	}
	w := bufio.NewWriter(tmp)
	var size int64
	for k, v := range s.data {
		n, err := w.Write(encode(opSet, k, v))
		size += int64(n)
		if err != nil {
			return cleanup(err)
		}
	}
	if err := w.Flush(); err != nil {
		return cleanup(err)
	}
	if err := tmp.Sync(); err != nil {
		return cleanup(err)
	}
	if err := os.Rename(tmpPath, filepath.Join(s.dir, LogFile)); err != nil {
		return cleanup(err)
	}
	syncDir(s.dir)
	s.f.Close()
	s.f = tmp // the renamed file; its offset is already at the end
	s.size = size
	return nil
}

// syncDir makes a rename durable on Linux. Some systems can't sync a folder, so the
// error is ignored.
func syncDir(dir string) {
	if d, err := os.Open(dir); err == nil {
		d.Sync()
		d.Close()
	}
}

// Close closes the log. Calling Close twice returns ErrClosed.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	s.closed = true
	return s.f.Close()
}
