// Package passwords stores and checks passenger passwords for the Twende matatu app:
// argon2id for new hashes, bcrypt hashes from the old system still accepted, and a
// rehash on login whenever a stored hash is weaker than today's settings.
package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
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
	n := utf8.RuneCountInString(password)
	if n < MinLength {
		return ErrTooShort
	}
	if n > MaxLength {
		return ErrTooLong
	}
	if blocklist[strings.ToLower(password)] {
		return ErrCommon
	}
	return nil
}

// HashArgon2id hashes password with a fresh random salt and returns it in the PHC
// string format: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>, base64 without padding.
func HashArgon2id(password string, p Params) (string, error) {
	salt := make([]byte, p.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// decodeArgon2id splits a PHC string into its parameters, salt and key.
func decodeArgon2id(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, key
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, ErrMalformedHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Params{}, nil, nil, ErrMalformedHash
	}
	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil {
		return Params{}, nil, nil, ErrMalformedHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return Params{}, nil, nil, ErrMalformedHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return Params{}, nil, nil, ErrMalformedHash
	}
	p.SaltLen = uint32(len(salt))
	p.KeyLen = uint32(len(key))
	return p, salt, key, nil
}

// VerifyArgon2id reports whether password matches the PHC string encoded. It
// recomputes the key with the stored salt and parameters and compares in constant time.
func VerifyArgon2id(encoded, password string) (bool, error) {
	p, salt, key, err := decodeArgon2id(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return subtle.ConstantTimeCompare(got, key) == 1, nil
}

// HashBcrypt hashes password with bcrypt at the given cost. The old Twende system
// stored these; new accounts use argon2id.
func HashBcrypt(password string, cost int) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// Verify checks password against a stored argon2id or bcrypt hash. needsRehash is true
// when the password matched but the hash should be replaced: any bcrypt hash, or an
// argon2id hash whose parameters differ from want.
func Verify(encoded, password string, want Params) (ok, needsRehash bool, err error) {
	switch {
	case strings.HasPrefix(encoded, "$argon2id$"):
		p, _, _, err := decodeArgon2id(encoded)
		if err != nil {
			return false, false, err
		}
		ok, err := VerifyArgon2id(encoded, password)
		if err != nil || !ok {
			return false, false, err
		}
		return true, p != want, nil
	case strings.HasPrefix(encoded, "$2a$"), strings.HasPrefix(encoded, "$2b$"), strings.HasPrefix(encoded, "$2y$"):
		err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, false, nil
		}
		if err != nil {
			return false, false, ErrMalformedHash
		}
		return true, true, nil
	default:
		return false, false, ErrMalformedHash
	}
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

	dummyOnce sync.Once
	dummy     string
}

// dummyHash is a real hash of a throwaway password, so a login for an unknown phone
// does the same slow work as one for a real account.
func (s *Service) dummyHash() string {
	s.dummyOnce.Do(func() {
		s.dummy, _ = HashArgon2id("not-a-real-password-just-padding", s.Params)
	})
	return s.dummy
}

// Register validates password, hashes it with argon2id and saves the user.
func (s *Service) Register(phone, password string) error {
	if _, ok := s.Store.Get(phone); ok {
		return ErrPhoneTaken
	}
	if err := ValidatePassword(password, s.Blocklist); err != nil {
		return err
	}
	hash, err := HashArgon2id(password, s.Params)
	if err != nil {
		return err
	}
	s.Store.Put(User{Phone: phone, PasswordHash: hash})
	s.Log.Info("user registered", "phone", phone)
	return nil
}

// Login checks phone and password. Unknown phones and wrong passwords both return
// ErrInvalidCredentials. When the stored hash needs it, Login saves a fresh argon2id hash.
func (s *Service) Login(phone, password string) error {
	u, found := s.Store.Get(phone)
	if !found {
		_, _, _ = Verify(s.dummyHash(), password, s.Params)
		s.Log.Warn("login failed", "phone", phone, "reason", "unknown phone")
		return ErrInvalidCredentials
	}
	ok, rehash, err := Verify(u.PasswordHash, password, s.Params)
	if err != nil {
		s.Log.Error("login failed", "phone", phone, "err", err)
		return ErrInvalidCredentials
	}
	if !ok {
		s.Log.Warn("login failed", "phone", phone, "reason", "wrong password")
		return ErrInvalidCredentials
	}
	if rehash {
		if h, err := HashArgon2id(password, s.Params); err == nil {
			u.PasswordHash = h
			s.Store.Put(u)
			s.Log.Info("password rehashed", "phone", phone)
		}
	}
	s.Log.Info("login ok", "phone", phone)
	return nil
}
