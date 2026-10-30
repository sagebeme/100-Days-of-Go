// Package proxy is a reverse proxy that forwards only to allowlisted upstreams,
// limits each client with a token bucket, and keeps its headers clean.
package proxy

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
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
	for _, v := range h.Values("Connection") {
		for _, name := range strings.Split(v, ",") {
			if name = strings.TrimSpace(name); name != "" {
				h.Del(name)
			}
		}
	}
	for _, name := range HopByHop {
		h.Del(name)
	}
}

// remoteIP is the IP the connection came from, without the port.
func remoteIP(r *http.Request) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	addr, err := netip.ParseAddr(host)
	return addr.Unmap(), err == nil
}

func trusted(addr netip.Addr, nets []netip.Prefix) bool {
	for _, p := range nets {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// ClientIP works out who the real client is. If the connection comes from a trusted
// proxy (a load balancer you run), it reads X-Forwarded-For from the right and returns
// the first address that isn't trusted. Otherwise it returns the connection's own
// address and ignores X-Forwarded-For, which any client can fake. If the header is
// missing, unreadable, or lists only trusted proxies, it returns the connection's address.
func ClientIP(r *http.Request, trustedProxies []netip.Prefix) string {
	peer, ok := remoteIP(r)
	if !ok {
		return r.RemoteAddr
	}
	if !trusted(peer, trustedProxies) {
		return peer.String()
	}
	var chain []string
	for _, v := range r.Header.Values("X-Forwarded-For") {
		chain = append(chain, strings.Split(v, ",")...)
	}
	for i := len(chain) - 1; i >= 0; i-- {
		addr, err := netip.ParseAddr(strings.TrimSpace(chain[i]))
		if err != nil {
			break // junk in the chain: don't trust anything to the left of it
		}
		addr = addr.Unmap()
		if !trusted(addr, trustedProxies) {
			return addr.String()
		}
	}
	return peer.String()
}
