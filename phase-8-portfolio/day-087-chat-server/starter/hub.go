// Package chat is a WebSocket chat server with rooms. One hub goroutine owns every
// room, so no locks are needed around them.
package chat

import "context"

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
	Name string
	Room string
	// Your fields here: the buffered channel, and an atomic count of dropped messages.
}

// NewClient returns a client whose channel holds up to buffer messages.
func NewClient(name, room string, buffer int) *Client {
	// Your code here
	return &Client{Name: name, Room: room}
}

// Messages is the channel the hub delivers to. The hub closes it when the client
// leaves, is disconnected for being slow, or the hub stops.
func (c *Client) Messages() <-chan Message {
	// Your code here
	return nil
}

// Dropped is how many messages this client missed under DropMessages.
func (c *Client) Dropped() int64 {
	// Your code here
	return 0
}

// Hub owns the rooms. Start it with Run; every other method talks to that goroutine
// over a channel, so calls are handled one at a time, in order.
type Hub struct {
	// Your fields here: the policy, and channels for join, leave, broadcast, queries
	// and "done".
}

// NewHub returns a hub that isn't running yet.
func NewHub(p Policy) *Hub {
	// Your code here
	return &Hub{}
}

// Run handles joins, leaves and broadcasts until ctx is cancelled. Then it closes
// every client's channel and returns. A join or leave is announced to the room with a
// TypeJoin or TypeLeave message. Delivery never blocks: a client whose buffer is full
// is handled by the policy. Empty rooms are deleted.
func (h *Hub) Run(ctx context.Context) {
	// Your code here
}

// Join adds c to its room and tells the room, c included. It returns false if the
// hub has stopped.
func (h *Hub) Join(c *Client) bool {
	// Your code here
	return false
}

// Leave removes c, closes its channel and tells the room. Leaving twice is safe.
func (h *Hub) Leave(c *Client) {
	// Your code here
}

// Broadcast sends m to everyone in m.Room, the sender included.
func (h *Hub) Broadcast(m Message) {
	// Your code here
}

// Stopped reports whether Run has returned or is shutting down.
func (h *Hub) Stopped() bool {
	// Your code here
	return false
}

// Rooms returns the names in each room, sorted. Because the hub handles requests
// in order, it reflects every Join, Leave and Broadcast that returned before it.
// It returns nil if the hub has stopped.
func (h *Hub) Rooms() map[string][]string {
	// Your code here
	return nil
}
