package chat

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// startHub runs a hub until the test ends.
func startHub(t *testing.T, p Policy) *Hub {
	t.Helper()
	h := NewHub(p)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		h.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return h
}

// next reads one message, failing the test if none arrives within a second.
func next(t *testing.T, c *Client) Message {
	t.Helper()
	select {
	case m, ok := <-c.Messages():
		if !ok {
			t.Fatalf("%s's channel was closed, want a message", c.Name)
		}
		return m
	case <-time.After(time.Second):
		t.Fatalf("%s got no message within a second", c.Name)
	}
	return Message{}
}

// closed reports whether c's channel is closed once its buffer is drained.
func closed(c *Client) bool {
	for {
		select {
		case _, ok := <-c.Messages():
			if !ok {
				return true
			}
		default:
			return false
		}
	}
}

func TestJoinAndRooms(t *testing.T) {
	h := startHub(t, Disconnect)
	otieno := NewClient("otieno", "route-125", 8)
	wanjiku := NewClient("wanjiku", "route-125", 8)
	kamau := NewClient("kamau", "route-46K", 8)
	for _, c := range []*Client{otieno, wanjiku, kamau} {
		if !h.Join(c) {
			t.Fatalf("Join(%s) = false on a running hub", c.Name)
		}
	}
	want := map[string][]string{"route-125": {"otieno", "wanjiku"}, "route-46K": {"kamau"}}
	if got := h.Rooms(); !reflect.DeepEqual(got, want) {
		t.Errorf("Rooms() = %v, want %v", got, want)
	}

	if m := next(t, otieno); m != (Message{Type: TypeJoin, Room: "route-125", From: "otieno"}) {
		t.Errorf("otieno's first message = %+v, want his own join", m)
	}
	if m := next(t, otieno); m != (Message{Type: TypeJoin, Room: "route-125", From: "wanjiku"}) {
		t.Errorf("otieno's second message = %+v, want wanjiku's join", m)
	}
	if m := next(t, kamau); m.From != "kamau" {
		t.Errorf("kamau saw %+v, want only his own join", m)
	}
}

func TestBroadcastStaysInRoom(t *testing.T) {
	h := startHub(t, Disconnect)
	a := NewClient("a", "route-125", 8)
	b := NewClient("b", "route-125", 8)
	c := NewClient("c", "route-111", 8)
	h.Join(a)
	h.Join(b)
	h.Join(c)
	next(t, a) // a joined
	next(t, a) // b joined
	next(t, b) // b joined
	next(t, c) // c joined

	sent := Message{Type: TypeMessage, Room: "route-125", From: "a", Text: "Mat iko Railways?"}
	h.Broadcast(sent)
	if got := next(t, a); got != sent {
		t.Errorf("sender got %+v, want %+v", got, sent)
	}
	if got := next(t, b); got != sent {
		t.Errorf("b got %+v, want %+v", got, sent)
	}
	h.Rooms() // the hub has finished the broadcast once this returns
	select {
	case m := <-c.Messages():
		t.Errorf("a client in another room got %+v", m)
	default:
	}
}

func TestLeave(t *testing.T) {
	h := startHub(t, Disconnect)
	a := NewClient("a", "r", 8)
	b := NewClient("b", "r", 8)
	h.Join(a)
	h.Join(b)
	h.Leave(a)
	h.Leave(a) // twice is safe
	if !closed(a) {
		t.Errorf("a's channel is still open after Leave")
	}
	next(t, b) // b joined
	if m := next(t, b); m != (Message{Type: TypeLeave, Room: "r", From: "a"}) {
		t.Errorf("b got %+v, want a's leave", m)
	}
	h.Leave(b)
	if got := h.Rooms(); len(got) != 0 {
		t.Errorf("Rooms() = %v after everyone left, want an empty map (no empty rooms)", got)
	}
}

func TestSlowClientDisconnected(t *testing.T) {
	h := startHub(t, Disconnect)
	slow := NewClient("slow", "r", 1)
	fast := NewClient("fast", "r", 16)
	h.Join(slow) // fills slow's buffer of 1 with its own join
	h.Join(fast) // no room for this one: slow is disconnected
	for i := range 3 {
		h.Broadcast(Message{Type: TypeMessage, Room: "r", From: "fast", Text: string(rune('x' + i))})
	}
	if got := h.Rooms()["r"]; !reflect.DeepEqual(got, []string{"fast"}) {
		t.Errorf("room r = %v, want only fast left", got)
	}
	if !closed(slow) {
		t.Errorf("the slow client's channel is still open")
	}
	if got := len(fast.Messages()); got != 4 {
		t.Errorf("fast has %d messages queued, want 4 (its join and 3 broadcasts)", got)
	}
}

func TestSlowClientDropped(t *testing.T) {
	h := startHub(t, DropMessages)
	slow := NewClient("slow", "r", 1)
	h.Join(slow)
	for range 3 {
		h.Broadcast(Message{Type: TypeMessage, Room: "r", From: "x", Text: "hi"})
	}
	h.Rooms()
	if got := slow.Dropped(); got != 3 {
		t.Errorf("Dropped() = %d, want 3", got)
	}
	if got := h.Rooms()["r"]; !reflect.DeepEqual(got, []string{"slow"}) {
		t.Errorf("room r = %v, want slow still in it", got)
	}
	if m := next(t, slow); m.Type != TypeJoin {
		t.Errorf("slow's queued message = %+v, want its join", m)
	}
}

func TestStop(t *testing.T) {
	h := NewHub(Disconnect)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		h.Run(ctx)
		close(done)
	}()
	a := NewClient("a", "r", 8)
	h.Join(a)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run didn't return after the context was cancelled")
	}
	if !closed(a) {
		t.Errorf("a's channel is still open after the hub stopped")
	}
	if h.Join(NewClient("b", "r", 1)) {
		t.Errorf("Join on a stopped hub = true, want false")
	}
	if got := h.Rooms(); got != nil {
		t.Errorf("Rooms on a stopped hub = %v, want nil", got)
	}
	h.Broadcast(Message{Room: "r"}) // must not block
	h.Leave(a)                      // must not block
}
