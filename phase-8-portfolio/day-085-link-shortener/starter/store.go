package linkshortener

import (
	"context"
	"errors"
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
	// Your fields here: a mutex and a map.
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	// Your code here
	return &MemoryStore{}
}

// Create implements Store.
func (m *MemoryStore) Create(ctx context.Context, l Link) error {
	// Your code here
	return nil
}

// Get implements Store.
func (m *MemoryStore) Get(ctx context.Context, code string) (Link, error) {
	// Your code here
	return Link{}, nil
}

// Hit implements Store.
func (m *MemoryStore) Hit(ctx context.Context, code string) (Link, error) {
	// Your code here
	return Link{}, nil
}
