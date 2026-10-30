// Package proxy is a reverse proxy that forwards only to allowlisted upstreams,
// limits each client with a token bucket, and keeps its headers clean.
package proxy

import (
	"net/http"
	"net/netip"
)

// HopByHop lists the headers that describe one connection, not the request. A proxy
// must not forward them (RFC 9110, section 7.6.1).
var HopByHop = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Proxy-Connection",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

// StripHopByHop removes the HopByHop headers from h, and every header named in its
// Connection header, such as "Connection: close, X-Debug".
func StripHopByHop(h http.Header) {
	// Your code here
}

// ClientIP works out who the real client is. If the connection comes from a trusted
// proxy (a load balancer you run), it reads X-Forwarded-For from the right and returns
// the first address that isn't trusted. Otherwise it returns the connection's own
// address and ignores X-Forwarded-For, which any client can fake. If the header is
// missing, unreadable, or lists only trusted proxies, it returns the connection's address.
func ClientIP(r *http.Request, trustedProxies []netip.Prefix) string {
	// Your code here
	return ""
}
