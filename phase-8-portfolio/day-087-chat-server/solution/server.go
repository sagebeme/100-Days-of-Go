package chat

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Limits on what a client may send.
const (
	MaxText      = 500     // characters in one message
	MaxFrame     = 4 << 10 // bytes in one WebSocket message
	SendBuffer   = 16      // messages queued for a client before it counts as slow
	WriteTimeout = 5 * time.Second
)

var validName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,20}$`)

// ValidName reports whether s can be a user or room name: 1 to 20 letters, digits,
// underscores or hyphens.
func ValidName(s string) bool { return validName.MatchString(s) }

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
	s.mux.HandleFunc("GET /ws", s.serveWS)
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	room, name := r.URL.Query().Get("room"), r.URL.Query().Get("name")
	if !ValidName(room) || !ValidName(name) {
		http.Error(w, "room and name must be 1 to 20 letters, digits, _ or -", http.StatusBadRequest)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.origins})
	if err != nil {
		return // Accept has already written the error response
	}
	defer conn.CloseNow()
	conn.SetReadLimit(MaxFrame)

	c := NewClient(name, room, SendBuffer)
	if !s.hub.Join(c) {
		conn.Close(websocket.StatusGoingAway, "server shutting down")
		return
	}
	defer s.hub.Leave(c)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go s.writeLoop(ctx, cancel, conn, c)

	for {
		var in struct {
			Text string `json:"text"`
		}
		if err := wsjson.Read(ctx, conn, &in); err != nil {
			return // closed by the client, too big, not JSON, or kicked by writeLoop
		}
		text := strings.TrimSpace(in.Text)
		if text == "" || utf8.RuneCountInString(text) > MaxText {
			continue
		}
		s.hub.Broadcast(Message{Type: TypeMessage, Room: room, From: name, Text: text})
	}
}

// writeLoop sends every message the hub delivers. When the hub closes the channel
// (a slow client, or shutdown) it closes the socket, which ends the read loop.
func (s *Server) writeLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, c *Client) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-c.Messages():
			if !ok {
				if s.hub.Stopped() {
					conn.Close(websocket.StatusGoingAway, "server shutting down")
				} else {
					conn.Close(websocket.StatusPolicyViolation, "too slow")
				}
				return
			}
			wctx, done := context.WithTimeout(ctx, WriteTimeout)
			err := wsjson.Write(wctx, conn, m)
			done()
			if err != nil {
				return
			}
		}
	}
}
