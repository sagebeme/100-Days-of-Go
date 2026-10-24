// Package chat is a WebSocket chat server with rooms. One hub goroutine owns every
// room, so no locks are needed around them.
package chat

import (
	"context"
	"slices"
	"sync/atomic"
)

// Message types.
const (
	TypeMessage = "message"
	TypeJoin    = "join"
	TypeLeave   = "leave"
)

// Message is what clients send and receive, as JSON.
type Message struct {
	Type string `json:"type"`
	Room string `json:"room"`
	From string `json:"from"`
	Text string `json:"text,omitempty"`
}

// Policy says what the hub does when a client's buffer is full.
type Policy int

const (
	// Disconnect removes a slow client from its room and closes its channel.
	Disconnect Policy = iota
	// DropMessages keeps the client but skips messages it has no room for.
	DropMessages
)

// Client is one connected user. The hub writes to its buffered channel and never
// waits for it.
type Client struct {
	Name    string
	Room    string
	send    chan Message
	dropped atomic.Int64
}

// NewClient returns a client whose channel holds up to buffer messages.
func NewClient(name, room string, buffer int) *Client {
	return &Client{Name: name, Room: room, send: make(chan Message, buffer)}
}

// Messages is the channel the hub delivers to. The hub closes it when the client
// leaves, is disconnected for being slow, or the hub stops.
func (c *Client) Messages() <-chan Message { return c.send }

// Dropped is how many messages this client missed under DropMessages.
func (c *Client) Dropped() int64 { return c.dropped.Load() }

// Hub owns the rooms. Start it with Run; every other method talks to that goroutine
// over a channel, so calls are handled one at a time, in order.
type Hub struct {
	policy    Policy
	join      chan *Client
	leave     chan *Client
	broadcast chan Message
	query     chan chan map[string][]string
	done      chan struct{}
}

// NewHub returns a hub that isn't running yet.
func NewHub(p Policy) *Hub {
	return &Hub{
		policy:    p,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		broadcast: make(chan Message),
		query:     make(chan chan map[string][]string),
		done:      make(chan struct{}),
	}
}

// Run handles joins, leaves and broadcasts until ctx is cancelled. Then it closes
// every client's channel and returns. A join or leave is announced to the room with a
// TypeJoin or TypeLeave message. Delivery never blocks: a client whose buffer is full
// is handled by the policy. Empty rooms are deleted.
func (h *Hub) Run(ctx context.Context) {
	rooms := map[string]map[*Client]bool{}

	remove := func(c *Client) {
		members := rooms[c.Room]
		if !members[c] {
			return // already gone: left twice, or kicked and then left
		}
		delete(members, c)
		close(c.send)
		if len(members) == 0 {
			delete(rooms, c.Room)
		}
	}
	deliver := func(m Message) {
		for c := range rooms[m.Room] {
			select {
			case c.send <- m:
			default:
				if h.policy == DropMessages {
					c.dropped.Add(1)
				} else {
					remove(c)
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			close(h.done) // first, so a closed client channel can tell it's a shutdown
			for _, members := range rooms {
				for c := range members {
					close(c.send)
				}
			}
			return
		case c := <-h.join:
			if rooms[c.Room] == nil {
				rooms[c.Room] = map[*Client]bool{}
			}
			rooms[c.Room][c] = true
			deliver(Message{Type: TypeJoin, Room: c.Room, From: c.Name})
		case c := <-h.leave:
			if rooms[c.Room][c] {
				remove(c)
				deliver(Message{Type: TypeLeave, Room: c.Room, From: c.Name})
			}
		case m := <-h.broadcast:
			deliver(m)
		case reply := <-h.query:
			out := map[string][]string{}
			for room, members := range rooms {
				for c := range members {
					out[room] = append(out[room], c.Name)
				}
				slices.Sort(out[room])
			}
			reply <- out
		}
	}
}

// Join adds c to its room and tells the room, c included. It returns false if the
// hub has stopped.
func (h *Hub) Join(c *Client) bool {
	select {
	case h.join <- c:
		return true
	case <-h.done:
		return false
	}
}

// Leave removes c, closes its channel and tells the room. Leaving twice is safe.
func (h *Hub) Leave(c *Client) {
	select {
	case h.leave <- c:
	case <-h.done:
	}
}

// Broadcast sends m to everyone in m.Room, the sender included.
func (h *Hub) Broadcast(m Message) {
	select {
	case h.broadcast <- m:
	case <-h.done:
	}
}

// Stopped reports whether Run has returned or is shutting down.
func (h *Hub) Stopped() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

// Rooms returns the names in each room, sorted. Because the hub handles requests
// in order, it reflects every Join, Leave and Broadcast that returned before it.
// It returns nil if the hub has stopped.
func (h *Hub) Rooms() map[string][]string {
	reply := make(chan map[string][]string, 1)
	select {
	case h.query <- reply:
		return <-reply
	case <-h.done:
		return nil
	}
}
