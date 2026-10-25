// Command api is Day 71's live seat map. Run it from the repo root:
//
//	go run ./phase-6-frontend/day-071-live-seats/api
//
// Connect a WebSocket to ws://localhost:8080/ws/seats. The server sends
//
//	{"type":"snapshot","version":7,"taken":[false,true,...]}   once, on connect (14 seats)
//	{"type":"seat","version":8,"seat":3,"taken":true}          whenever a seat changes
//	{"type":"error","message":"Seat 3 is already taken"}       when your hold fails
//
// and accepts {"type":"hold","seat":3}. Seats are numbered from 1.
//
// Environment: ADDR (default :8080), ALLOWED_ORIGIN (default http://localhost:5173),
// SIMULATE (default on: other "passengers" take and free seats every few seconds; set off to stop).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const seatCount = 14

// hub owns the seat map and everyone watching it.
type hub struct {
	mu      sync.Mutex
	taken   [seatCount]bool
	version int
	subs    map[chan []byte]bool
}

func newHub() *hub { return &hub{subs: map[chan []byte]bool{}} }

// subscribe returns the whole map as JSON, and a channel that will get every change after it.
func (h *hub) subscribe() ([]byte, chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 16)
	h.subs[ch] = true
	b, _ := json.Marshal(map[string]any{"type": "snapshot", "version": h.version, "taken": h.taken[:]})
	return b, ch
}

func (h *hub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, ch)
}

// set changes one seat and tells every subscriber. It reports false if the seat was already in that state.
func (h *hub) set(seat int, taken bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if seat < 1 || seat > seatCount || h.taken[seat-1] == taken {
		return false
	}
	h.taken[seat-1] = taken
	h.version++
	b, _ := json.Marshal(map[string]any{"type": "seat", "version": h.version, "seat": seat, "taken": taken})
	for ch := range h.subs {
		select {
		case ch <- b:
		default: // this client is too slow to keep up: drop it, and it will reconnect for a fresh snapshot
			delete(h.subs, ch)
			close(ch)
		}
	}
	return true
}

func main() {
	addr := getenv("ADDR", ":8080")
	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	h := newHub()
	if getenv("SIMULATE", "on") != "off" {
		go simulate(h)
	}
	log.Printf("listening on %s, allowing origin %s", addr, origin)
	log.Fatal(http.ListenAndServe(addr, h.routes(origin)))
}

func (h *hub) routes(origin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/seats", func(w http.ResponseWriter, r *http.Request) {
		// Browsers don't apply CORS to WebSockets, so the server must check Origin itself.
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{origin}})
		if err != nil {
			return // Accept has already written the error response
		}
		defer c.CloseNow()
		h.serve(r.Context(), c)
	})
	return mux
}

func (h *hub) serve(ctx context.Context, c *websocket.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	snap, updates := h.subscribe()
	defer h.unsubscribe(updates)

	// Reader: holds from this passenger. It cancels ctx when the connection closes.
	errs := make(chan string, 4)
	go func() {
		defer cancel()
		for {
			var in struct {
				Type string `json:"type"`
				Seat int    `json:"seat"`
			}
			if err := wsjson.Read(ctx, c, &in); err != nil {
				return
			}
			if in.Type == "hold" && !h.set(in.Seat, true) {
				select {
				case errs <- fmt.Sprintf("Seat %d is already taken", in.Seat):
				default:
				}
			}
		}
	}()

	// Writer: only this goroutine writes, so writes never interleave.
	if c.Write(ctx, websocket.MessageText, snap) != nil {
		return
	}
	for {
		select {
		case b, ok := <-updates:
			if !ok {
				c.Close(websocket.StatusTryAgainLater, "too slow")
				return
			}
			if c.Write(ctx, websocket.MessageText, b) != nil {
				return
			}
		case msg := <-errs:
			if wsjson.Write(ctx, c, map[string]string{"type": "error", "message": msg}) != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// simulate plays other passengers, so you can watch seats change live.
func simulate(h *hub) {
	for range time.Tick(3 * time.Second) {
		seat := rand.IntN(seatCount) + 1
		h.mu.Lock()
		taken := h.taken[seat-1]
		h.mu.Unlock()
		h.set(seat, !taken)
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
