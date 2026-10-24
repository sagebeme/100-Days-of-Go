package linkshortener

import (
	"context"
	"errors"
	"sync"
)

// ErrNotFound means no link has that code.
var ErrNotFound = errors.New("link not found")

// ErrCodeTaken means a link with that code already exists.
var ErrCodeTaken = errors.New("code already taken")

// Store keeps links. The API depends only on this interface, so a Postgres or Redis
// store can replace MemoryStore without touching a handler.
type Store interface {
	// Create saves l. It returns ErrCodeTaken if l.Code is in use and never overwrites.
	Create(ctx context.Context, l Link) error
	// Get returns the link with that code, or ErrNotFound.
	Get(ctx context.Context, code string) (Link, error)
	// Hit adds one to the link's hit count and returns the link, or ErrNotFound.
	Hit(ctx context.Context, code string) (Link, error)
}

// MemoryStore is a Store in a map, safe for concurrent use. The zero value is not
// ready: use NewMemoryStore.
type MemoryStore struct {
	mu    sync.Mutex
	links map[string]Link
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: map[string]Link{}}
}

// Create implements Store.
func (m *MemoryStore) Create(_ context.Context, l Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[l.Code]; ok {
		return ErrCodeTaken
	}
	m.links[l.Code] = l
	return nil
}

// Get implements Store.
func (m *MemoryStore) Get(_ context.Context, code string) (Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return l, nil
}

// Hit implements Store.
func (m *MemoryStore) Hit(_ context.Context, code string) (Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	l.Hits++
	m.links[code] = l
	return l, nil
}
