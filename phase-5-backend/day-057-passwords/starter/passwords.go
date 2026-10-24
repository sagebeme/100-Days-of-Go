// Package passwords stores and checks passenger passwords for the Twende matatu app:
// argon2id for new hashes, bcrypt hashes from the old system still accepted, and a
// rehash on login whenever a stored hash is weaker than today's settings.
package passwords

import (
	"errors"
	"log/slog"
	"sync"
)

// MinLength and MaxLength are the password length rules, counted in characters (runes).
// NIST SP 800-63B-4 asks for at least 15 characters when the password is the only
// factor, and for verifiers to allow at least 64.
const (
	MinLength = 15
	MaxLength = 64
)

// Errors returned by this package. Callers compare them with errors.Is.
var (
	ErrTooShort           = errors.New("password must be at least 15 characters")
	ErrTooLong            = errors.New("password must be at most 64 characters")
	ErrCommon             = errors.New("password is too common, pick another")
	ErrMalformedHash      = errors.New("stored hash is malformed")
	ErrInvalidCredentials = errors.New("wrong phone number or password")
	ErrPhoneTaken         = errors.New("phone number already has an account")
)

// Params are the argon2id cost settings. Memory is in KiB.
type Params struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	SaltLen uint32
	KeyLen  uint32
}

// DefaultParams follow the OWASP Password Storage Cheat Sheet minimum for argon2id:
// 19 MiB of memory, 2 passes, 1 thread.
var DefaultParams = Params{Memory: 19 * 1024, Time: 2, Threads: 1, SaltLen: 16, KeyLen: 32}

// ValidatePassword applies length-over-composition rules: 15 to 64 characters, not on
// the blocklist (compared case-insensitively). It never demands digits or symbols.
func ValidatePassword(password string, blocklist map[string]bool) error {
	// Your code here
	return nil
}

// HashArgon2id hashes password with a fresh random salt and returns it in the PHC
// string format: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>, base64 without padding.
func HashArgon2id(password string, p Params) (string, error) {
	// Your code here
	return "", nil
}

// VerifyArgon2id reports whether password matches the PHC string encoded. It
// recomputes the key with the stored salt and parameters and compares in constant time.
// A string it can't parse returns ErrMalformedHash.
func VerifyArgon2id(encoded, password string) (bool, error) {
	// Your code here
	return false, nil
}

// HashBcrypt hashes password with bcrypt at the given cost. The old Twende system
// stored these; new accounts use argon2id.
func HashBcrypt(password string, cost int) (string, error) {
	// Your code here
	return "", nil
}

// Verify checks password against a stored argon2id or bcrypt hash. needsRehash is true
// when the password matched but the hash should be replaced: any bcrypt hash, or an
// argon2id hash whose parameters differ from want. Any other format is ErrMalformedHash.
func Verify(encoded, password string, want Params) (ok, needsRehash bool, err error) {
	// Your code here
	return false, false, nil
}

// User is a passenger account. PasswordHash is the only form of the password kept.
type User struct {
	Phone        string
	PasswordHash string
}

// Store keeps users by phone number.
type Store interface {
	Get(phone string) (User, bool)
	Put(u User)
}

// MemStore is an in-memory Store, safe for concurrent use.
type MemStore struct {
	mu    sync.Mutex
	users map[string]User
}

// NewMemStore returns an empty MemStore.
func NewMemStore() *MemStore { return &MemStore{users: map[string]User{}} }

// Get returns the user with this phone number.
func (s *MemStore) Get(phone string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[phone]
	return u, ok
}

// Put saves u, replacing any user with the same phone number.
func (s *MemStore) Put(u User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.Phone] = u
}

// Service registers and logs in passengers.
type Service struct {
	Store     Store
	Params    Params
	Blocklist map[string]bool
	Log       *slog.Logger
}

// Register validates password, hashes it with argon2id and saves the user.
// A phone number that already has an account returns ErrPhoneTaken.
func (s *Service) Register(phone, password string) error {
	// Your code here
	return nil
}

// Login checks phone and password. Unknown phones and wrong passwords both return
// ErrInvalidCredentials. When the stored hash needs it, Login saves a fresh argon2id hash.
func (s *Service) Login(phone, password string) error {
	// Your code here
	return nil
}
