package chat

import (
	"net/http"
	"time"
)

// Limits on what a client may send.
const (
	MaxText      = 500     // characters in one message
	MaxFrame     = 4 << 10 // bytes in one WebSocket message
	SendBuffer   = 16      // messages queued for a client before it counts as slow
	WriteTimeout = 5 * time.Second
)

// ValidName reports whether s can be a user or room name: 1 to 20 letters, digits,
// underscores or hyphens.
func ValidName(s string) bool {
	// Your code here
	return false
}

// Server upgrades GET /ws?room=...&name=... to a WebSocket and connects it to a hub.
// Clients send {"text":"..."}; the server fills in the room and the name itself.
type Server struct {
	hub     *Hub
	origins []string
	mux     *http.ServeMux
}

// NewServer returns a server for hub. origins lists the other hosts whose pages may
// open a socket (the server's own host is always allowed). Leave it empty unless a
// front end on another domain needs it.
func NewServer(hub *Hub, origins ...string) *Server {
	s := &Server{hub: hub, origins: origins, mux: http.NewServeMux()}
	// Your code here: register GET /ws.
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
