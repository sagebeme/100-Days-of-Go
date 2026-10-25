package secretsmanagement

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
)

// Key is one signing key. ID is public and travels with each signature; Secret never
// leaves the server.
type Key struct {
	ID     string
	Secret Secret
}

// ErrBadSignature is returned by Verify for anything it can't accept.
var ErrBadSignature = errors.New("bad signature")

// Keyring signs with its current key and verifies with the current or the previous
// one. That's how keys rotate without breaking anything: add a new key, keep
// accepting the old one until everything it signed has expired, then drop it.
type Keyring struct {
	mu       sync.RWMutex
	current  Key
	previous *Key
}

// NewKeyring returns a Keyring that signs with k.
func NewKeyring(k Key) *Keyring {
	return &Keyring{current: k}
}

// Rotate makes k the current key. The old current key becomes the previous one and
// the old previous key is forgotten.
func (r *Keyring) Rotate(k Key) {
	// Your code here
}

// Sign returns "<key id>.<signature>": an HMAC-SHA256 of msg with the current key,
// base64url-encoded without padding.
//
//	Sign("booking:1042") → "k2.Xb3..."
func (r *Keyring) Sign(msg string) string {
	// Your code here
	return ""
}

// Verify checks a signature from Sign. It picks the key by the ID in front of the dot
// (current or previous) and compares MACs in constant time with hmac.Equal. An
// unknown ID, a missing dot or a wrong MAC is ErrBadSignature.
func (r *Keyring) Verify(msg, sig string) error {
	// Your code here
	return nil
}

func mac(k Key, msg string) string {
	h := hmac.New(sha256.New, []byte(k.Secret.Reveal()))
	h.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
