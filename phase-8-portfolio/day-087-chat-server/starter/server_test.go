package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestValidName(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"otieno", true},
		{"route-125", true},
		{"Mama_Mboga2", true},
		{"", false},
		{"two words", false},
		{"<script>", false},
		{strings.Repeat("a", 21), false},
	}
	for _, tt := range tests {
		if got := ValidName(tt.s); got != tt.want {
			t.Errorf("ValidName(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

// dial connects to the test server, failing the test on error.
func dial(t *testing.T, ctx context.Context, srv *httptest.Server, room, name string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?room=" + room + "&name=" + name
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial %s/%s: %v", room, name, err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

// waitFor reads until a message matching ok arrives, and returns it.
func waitFor(t *testing.T, ctx context.Context, conn *websocket.Conn, ok func(Message) bool) Message {
	t.Helper()
	for {
		var m Message
		if err := wsjson.Read(ctx, conn, &m); err != nil {
			t.Fatalf("read: %v", err)
		}
		if ok(m) {
			return m
		}
	}
}

func chatText(m Message) bool { return m.Type == TypeMessage }

func TestChatOverWebSockets(t *testing.T) {
	h := startHub(t, Disconnect)
	srv := httptest.NewServer(NewServer(h))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	otieno := dial(t, ctx, srv, "route-125", "otieno")
	waitFor(t, ctx, otieno, func(m Message) bool { return m.Type == TypeJoin && m.From == "otieno" })
	wanjiku := dial(t, ctx, srv, "route-125", "wanjiku")
	waitFor(t, ctx, otieno, func(m Message) bool { return m.Type == TypeJoin && m.From == "wanjiku" })
	kamau := dial(t, ctx, srv, "route-46K", "kamau")
	waitFor(t, ctx, kamau, func(m Message) bool { return m.Type == TypeJoin && m.From == "kamau" })

	if err := wsjson.Write(ctx, otieno, map[string]string{"text": "  Mat ya 125 imefika Railways  "}); err != nil {
		t.Fatalf("write: %v", err)
	}
	want := Message{Type: TypeMessage, Room: "route-125", From: "otieno", Text: "Mat ya 125 imefika Railways"}
	if got := waitFor(t, ctx, wanjiku, chatText); got != want {
		t.Errorf("wanjiku got %+v, want %+v", got, want)
	}
	if got := waitFor(t, ctx, otieno, chatText); got != want {
		t.Errorf("the sender got %+v, want its own message back", got)
	}

	// Empty and over-long messages are ignored, and the name can't be faked.
	wsjson.Write(ctx, wanjiku, map[string]string{"text": "   "})
	wsjson.Write(ctx, wanjiku, map[string]string{"text": strings.Repeat("a", MaxText+1)})
	wsjson.Write(ctx, wanjiku, map[string]string{"text": "Sawa", "from": "otieno"})
	if got := waitFor(t, ctx, otieno, chatText); got.Text != "Sawa" || got.From != "wanjiku" {
		t.Errorf("otieno got %+v, want Sawa from wanjiku and nothing before it", got)
	}

	// kamau's room never saw route-125's messages: his next chat message is his own.
	wsjson.Write(ctx, kamau, map[string]string{"text": "Kawangware sasa"})
	if got := waitFor(t, ctx, kamau, chatText); got.Text != "Kawangware sasa" {
		t.Errorf("kamau got %+v from another room", got)
	}

	// When wanjiku hangs up, otieno is told.
	wanjiku.Close(websocket.StatusNormalClosure, "")
	if got := waitFor(t, ctx, otieno, func(m Message) bool { return m.Type == TypeLeave }); got.From != "wanjiku" {
		t.Errorf("leave message = %+v, want wanjiku", got)
	}
}

func TestBadRequests(t *testing.T) {
	h := startHub(t, Disconnect)
	srv := httptest.NewServer(NewServer(h))
	t.Cleanup(srv.Close)

	for _, q := range []string{"", "?room=r", "?name=n", "?room=r&name=two%20words", "?room=%3Cb%3E&name=n"} {
		res, err := http.Get(srv.URL + "/ws" + q)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("GET /ws%s status = %d, want 400", q, res.StatusCode)
		}
	}

	// A browser page on another site may not open a socket (cross-site WebSocket hijacking).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?room=r&name=n"
	_, res, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": {"https://evil.example"}},
	})
	if err == nil {
		t.Fatalf("dial with a foreign Origin succeeded, want it refused")
	}
	if res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("foreign Origin response = %v, want 403", res)
	}
}

func TestServerClosesOnShutdown(t *testing.T) {
	h := NewHub(Disconnect)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		h.Run(ctx)
		close(stopped)
	}()
	srv := httptest.NewServer(NewServer(h))
	t.Cleanup(srv.Close)

	rctx, rcancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer rcancel()
	conn := dial(t, rctx, srv, "r", "n")
	waitFor(t, rctx, conn, func(m Message) bool { return m.Type == TypeJoin })

	cancel() // stop the hub: the server must close the socket
	<-stopped
	var m Message
	err := wsjson.Read(rctx, conn, &m)
	if websocket.CloseStatus(err) != websocket.StatusGoingAway {
		t.Errorf("read after shutdown: %v, want a close with StatusGoingAway", err)
	}
}
