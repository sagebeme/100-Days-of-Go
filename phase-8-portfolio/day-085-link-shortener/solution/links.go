// Package linkshortener is a small URL shortener: base62 codes, a store behind an
// interface, URL checks that refuse unsafe targets, and an HTTP API.
package linkshortener

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
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
	if n == 0 {
		return "0"
	}
	var buf [11]byte // 62^11 > 2^64
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = Alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

// Decode reads a base62 string back into a number. It returns ErrBadCode for an
// empty string, a character outside Alphabet, or a value that overflows uint64.
func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrBadCode
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		d := strings.IndexByte(Alphabet, s[i])
		if d < 0 {
			return 0, ErrBadCode
		}
		if n > (^uint64(0)-uint64(d))/62 {
			return 0, ErrBadCode
		}
		n = n*62 + uint64(d)
	}
	return n, nil
}

// ValidCode reports whether s looks like a code this service could have made:
// 1 to 16 base62 characters.
func ValidCode(s string) bool {
	if len(s) == 0 || len(s) > 16 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(Alphabet, s[i]) < 0 {
			return false
		}
	}
	return true
}

// RandomCode returns CodeLength random base62 characters from crypto/rand.
func RandomCode() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // crypto/rand.Read never fails on supported platforms
	}
	const space = 3521614606208 // 62^7
	code := Encode(binary.BigEndian.Uint64(b[:]) % space)
	return strings.Repeat("0", CodeLength-len(code)) + code
}

// LookupFunc resolves a host name to IP addresses. net.DefaultResolver.LookupNetIP
// fits, and tests pass a fake.
type LookupFunc func(ctx context.Context, host string) ([]netip.Addr, error)

// DefaultLookup resolves with the system resolver.
func DefaultLookup(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

// PublicAddr reports whether addr is a public unicast address: not loopback,
// private (10/8, 172.16/12, 192.168/16, fc00::/7), link-local (169.254/16, which holds
// the cloud metadata address, and fe80::/10), carrier-grade NAT (100.64/10),
// unspecified, multicast or broadcast.
func PublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap() // ::ffff:127.0.0.1 is 127.0.0.1
	if !addr.IsValid() || addr.IsLoopback() || addr.IsPrivate() || addr.IsUnspecified() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() ||
		addr.IsInterfaceLocalMulticast() {
		return false
	}
	if cgnat.Contains(addr) || addr == netip.AddrFrom4([4]byte{255, 255, 255, 255}) {
		return false
	}
	return true
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

// ValidateURL checks a long URL before it's stored. It must be at most MaxURLLength
// bytes, absolute, http or https, with a host and no user:password part. The host
// must not be localhost, and every address it is or resolves to must be public.
// lookup resolves names; an IP literal is checked without it. A refusal is a *URLError.
func ValidateURL(ctx context.Context, raw string, lookup LookupFunc) (*url.URL, error) {
	if len(raw) > MaxURLLength {
		return nil, &URLError{Reason: fmt.Sprintf("longer than %d bytes", MaxURLLength)}
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, &URLError{Reason: "can't be parsed"}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, &URLError{Reason: "scheme must be http or https"}
	}
	if u.User != nil {
		return nil, &URLError{Reason: "must not contain a user name or password"}
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" {
		return nil, &URLError{Reason: "must have a host"}
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, &URLError{Reason: "must not point at this machine"}
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		if !PublicAddr(addr) {
			return nil, &URLError{Reason: "must not point at a private or local address"}
		}
		return u, nil
	}
	addrs, err := lookup(ctx, host)
	if err != nil || len(addrs) == 0 {
		return nil, &URLError{Reason: "host doesn't resolve"}
	}
	for _, a := range addrs {
		if !PublicAddr(a) {
			return nil, &URLError{Reason: "must not point at a private or local address"}
		}
	}
	return u, nil
}
