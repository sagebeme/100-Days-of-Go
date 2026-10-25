// Package secretsmanagement loads Twende's secrets from env variables, mounted files or
// a vault, keeps them out of logs, and rotates signing keys without logging anyone out.
package secretsmanagement

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"sync"
)

// Errors every Provider uses.
var (
	ErrNotFound  = errors.New("secret not found")
	ErrForbidden = errors.New("not allowed to read this secret")
	ErrBadName   = errors.New("bad secret name")
)

// Secret is a string that refuses to be printed. fmt and slog show "[REDACTED]";
// only Reveal returns the real value, so every use of it is easy to find in review.
type Secret string

// Redacted is what a Secret prints as.
const Redacted = "[REDACTED]"

// Reveal returns the real value.
func (s Secret) Reveal() string { return string(s) }

// String makes fmt's %s and %v print Redacted.
func (s Secret) String() string {
	// Your code here
	return ""
}

// GoString makes fmt's %#v print Redacted too.
func (s Secret) GoString() string {
	// Your code here
	return ""
}

// LogValue makes slog log Redacted, whichever handler is used.
func (s Secret) LogValue() slog.Value {
	// Your code here
	return slog.Value{}
}

// Provider looks a secret up by name.
type Provider interface {
	Get(ctx context.Context, name string) (Secret, error)
}

// Env reads secrets from env variables named Prefix + name, such as TWENDE_DB_PASSWORD
// for "DB_PASSWORD". Lookup defaults to os.LookupEnv; tests pass their own. An empty
// value counts as not set.
type Env struct {
	Prefix string
	Lookup func(string) (string, bool)
}

// Get implements Provider. It wraps ErrNotFound with the variable's name.
func (e Env) Get(ctx context.Context, name string) (Secret, error) {
	// Your code here
	return "", nil
}

// Files reads secrets from files, one per secret, the way Docker (/run/secrets) and
// Kubernetes (a mounted Secret volume) provide them. The file's name is the secret's
// name. One trailing newline ("\n" or "\r\n") is removed, because editors and `echo`
// add one.
type Files struct {
	FS fs.FS // os.DirFS("/run/secrets") in production, fstest.MapFS in tests
}

// Get implements Provider. A name that isn't a single plain file name (it contains a
// slash, or is "." or "..") is ErrBadName, so nobody can read ../../etc/passwd.
// A missing file is ErrNotFound.
func (f Files) Get(ctx context.Context, name string) (Secret, error) {
	// Your code here
	return "", nil
}

// Chain asks each provider in turn and returns the first secret found. Any error other
// than ErrNotFound stops the search, because a forbidden or broken source shouldn't
// silently fall through to an older copy somewhere else.
type Chain []Provider

// Get implements Provider.
func (c Chain) Get(ctx context.Context, name string) (Secret, error) {
	// Your code here
	return "", nil
}

// FakeVault is an in-memory stand-in for a secrets vault such as HashiCorp Vault or a
// cloud secrets manager. It keeps every version of each secret and gives each caller
// a token that may read only some paths.
type FakeVault struct {
	mu       sync.Mutex
	versions map[string][]string // path → values, oldest first
	policies map[string][]string // token → path patterns it may read
	audit    []string            // "token path" for every read attempt
}

// NewFakeVault returns an empty vault.
func NewFakeVault() *FakeVault {
	return &FakeVault{versions: map[string][]string{}, policies: map[string][]string{}}
}

// Put stores a new version of the secret at path and returns its version number,
// starting at 1. Older versions are kept, so a bad rotation can be rolled back.
func (v *FakeVault) Put(path, value string) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.versions[path] = append(v.versions[path], value)
	return len(v.versions[path])
}

// Grant lets token read paths matching the patterns. A pattern is an exact path, or
// a prefix ending in "/*" that matches everything under it: "twende/booking-api/*".
func (v *FakeVault) Grant(token string, patterns ...string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.policies[token] = append(v.policies[token], patterns...)
}

// allowed reports whether one of patterns matches path.
func allowed(patterns []string, path string) bool {
	// Your code here
	return false
}

// Read returns the latest version of the secret at path for token. Every attempt,
// allowed or not, is added to the audit log. A token without a matching grant gets
// ErrForbidden, even for a path that doesn't exist, so it can't probe what's there.
func (v *FakeVault) Read(token, path string) (Secret, error) {
	// Your code here
	return "", nil
}

// Audit returns a copy of the audit log.
func (v *FakeVault) Audit() []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return append([]string(nil), v.audit...)
}

// Client returns a Provider that reads from the vault as token, under prefix:
// Get(ctx, "DB_PASSWORD") reads prefix + "/DB_PASSWORD".
func (v *FakeVault) Client(token, prefix string) Provider {
	return vaultClient{v: v, token: token, prefix: prefix}
}

type vaultClient struct {
	v             *FakeVault
	token, prefix string
}

func (c vaultClient) Get(ctx context.Context, name string) (Secret, error) {
	return c.v.Read(c.token, c.prefix+"/"+name)
}
