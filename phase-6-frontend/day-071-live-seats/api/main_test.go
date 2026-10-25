package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type msg struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Taken   any    `json:"taken"`
	Seat    int    `json:"seat"`
	Message string `json:"message"`
}

func TestSeats(t *testing.T) {
	h := newHub()
	srv := httptest.NewServer(h.routes("http://localhost:5173"))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	url := "ws" + srv.URL[len("http"):] + "/ws/seats"
	dial := func() *websocket.Conn {
		c, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"http://localhost:5173"}}})
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		return c
	}
	read := func(c *websocket.Conn) msg {
		var m msg
		if err := wsjson.Read(ctx, c, &m); err != nil {
			t.Fatalf("read: %v", err)
		}
		return m
	}

	a, b := dial(), dial()
	defer a.CloseNow()
	defer b.CloseNow()
	if m := read(a); m.Type != "snapshot" || m.Version != 0 {
		t.Fatalf("first message = %+v, want a version 0 snapshot", m)
	}
	read(b)

	wsjson.Write(ctx, a, map[string]any{"type": "hold", "seat": 3})
	for _, c := range []*websocket.Conn{a, b} {
		if m := read(c); m.Type != "seat" || m.Seat != 3 || m.Taken != true || m.Version != 1 {
			t.Errorf("after hold: got %+v, want seat 3 taken at version 1", m)
		}
	}
	wsjson.Write(ctx, b, map[string]any{"type": "hold", "seat": 3})
	if m := read(b); m.Type != "error" || m.Message != "Seat 3 is already taken" {
		t.Errorf("second hold: got %+v, want an error", m)
	}
}

func TestOriginChecked(t *testing.T) {
	srv := httptest.NewServer(newHub().routes("http://localhost:5173"))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, res, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):]+"/ws/seats",
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})
	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("dial from another origin: err %v, response %v; want 403", err, res)
	}
}
