package owasptop10

import (
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

// ErrBlocked is returned for a URL or address the server must not fetch.
var ErrBlocked = errors.New("destination not allowed")

// extraBlocked are ranges netip has no method for: "this network" and the shared
// address space some clouds and carriers use internally.
var extraBlocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
}

// IsBlockedIP reports whether ip is somewhere the server must never connect to on a
// user's behalf: loopback, private, link-local (which holds the cloud metadata address
// 169.254.169.254), unspecified, multicast, or one of extraBlocked. IPv4 addresses
// written as IPv6 (::ffff:127.0.0.1) are unwrapped first.
func IsBlockedIP(ip netip.Addr) bool {
	// Your code here
	return false
}

// CheckURL parses raw and allows it only if it is https, has no user:password part,
// and its host is in allowed.
func CheckURL(raw string, allowed map[string]bool) (*url.URL, error) {
	// Your code here
	return nil, nil
}

// NewSafeClient returns an http.Client for fetching user-supplied URLs. Its dialer
// refuses blocked IPs after DNS lookup, it follows at most 3 redirects and only to
// URLs CheckURL allows, and every request times out after timeout.
func NewSafeClient(allowed map[string]bool, timeout time.Duration) *http.Client {
	// Your code here
	return &http.Client{}
}
