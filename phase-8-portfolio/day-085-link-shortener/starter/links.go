// Package linkshortener is a small URL shortener: base62 codes, a store behind an
// interface, URL checks that refuse unsafe targets, and an HTTP API.
package linkshortener

import (
	"context"
	"errors"
	"net/netip"
	"net/url"
	"time"
)

// Alphabet is the base62 alphabet: digits, then upper case, then lower case.
const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// CodeLength is the length of a generated short code. 62^7 is about 3.5 trillion.
const CodeLength = 7

// MaxURLLength is the longest URL the service accepts.
const MaxURLLength = 2048

// Link is one short link.
type Link struct {
	Code    string    `json:"code"`
	URL     string    `json:"url"`
	Hits    int64     `json:"hits"`
	Created time.Time `json:"created"`
}

// ErrBadCode means a string isn't valid base62.
var ErrBadCode = errors.New("not a base62 code")

// URLError says why a URL was refused. Its message is safe to show to the client.
type URLError struct {
	Reason string
}

func (e *URLError) Error() string { return "invalid url: " + e.Reason }

// Encode writes n in base62. Encode(0) is "0".
func Encode(n uint64) string {
	// Your code here
	return ""
}

// Decode reads a base62 string back into a number. It returns ErrBadCode for an
// empty string, a character outside Alphabet, or a value that overflows uint64.
func Decode(s string) (uint64, error) {
	// Your code here
	return 0, nil
}

// ValidCode reports whether s looks like a code this service could have made:
// 1 to 16 base62 characters.
func ValidCode(s string) bool {
	// Your code here
	return false
}

// RandomCode returns CodeLength random base62 characters from crypto/rand.
func RandomCode() string {
	// Your code here
	return ""
}

// LookupFunc resolves a host name to IP addresses. net.DefaultResolver.LookupNetIP
// fits, and tests pass a fake.
type LookupFunc func(ctx context.Context, host string) ([]netip.Addr, error)

// DefaultLookup resolves with the system resolver.
func DefaultLookup(ctx context.Context, host string) ([]netip.Addr, error) {
	// Your code here
	return nil, nil
}

// PublicAddr reports whether addr is a public unicast address: not loopback,
// private (10/8, 172.16/12, 192.168/16, fc00::/7), link-local (169.254/16, which holds
// the cloud metadata address, and fe80::/10), carrier-grade NAT (100.64/10),
// unspecified, multicast or broadcast.
func PublicAddr(addr netip.Addr) bool {
	// Your code here
	return false
}

// ValidateURL checks a long URL before it's stored. It must be at most MaxURLLength
// bytes, absolute, http or https, with a host and no user:password part. The host
// must not be localhost, and every address it is or resolves to must be public.
// lookup resolves names; an IP literal is checked without it. A refusal is a *URLError.
func ValidateURL(ctx context.Context, raw string, lookup LookupFunc) (*url.URL, error) {
	// Your code here
	return nil, nil
}
