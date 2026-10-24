// Command chat runs the chat server with a small test page at /.
//
//	go run ./phase-8-portfolio/day-087-chat-server/solution/cmd/chat -addr :8080
//
// Open http://localhost:8080/?room=route-125&name=otieno in two browser tabs.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	chat "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-087-chat-server/solution"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	drop := flag.Bool("drop", false, "drop messages for slow clients instead of disconnecting them")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	policy := chat.Disconnect
	if *drop {
		policy = chat.DropMessages
	}
	hub := chat.NewHub(policy)
	go hub.Run(ctx)

	mux := http.NewServeMux()
	mux.Handle("GET /ws", chat.NewServer(hub))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	})
	srv := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shut)
	}()
	log.Printf("chat listening on %s", *addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// page is a bare test client. It uses textContent, never innerHTML, so a message
// can't inject HTML.
const page = `<!doctype html>
<meta charset="utf-8">
<title>Chat</title>
<ul id="log"></ul>
<form id="f"><input id="t" autocomplete="off" maxlength="500"><button>Tuma</button></form>
<script>
const q = new URLSearchParams(location.search);
const ws = new WebSocket("ws://" + location.host + "/ws?room=" + encodeURIComponent(q.get("room") || "lobby") + "&name=" + encodeURIComponent(q.get("name") || "mgeni"));
const add = (s) => { const li = document.createElement("li"); li.textContent = s; document.getElementById("log").append(li); };
ws.onmessage = (e) => { const m = JSON.parse(e.data); add(m.type === "message" ? m.from + ": " + m.text : m.from + " " + m.type); };
ws.onclose = (e) => add("closed: " + e.code + " " + e.reason);
document.getElementById("f").onsubmit = (e) => { e.preventDefault(); const t = document.getElementById("t"); ws.send(JSON.stringify({text: t.value})); t.value = ""; };
</script>
`
