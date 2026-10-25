// Command katabot serves the kata bot on Discord's interactions endpoint.
//
//	DISCORD_PUBLIC_KEY=<hex from the developer portal> PORT=8080 go run ./phase-8-portfolio/day-096-discord-bot/solution/cmd/katabot
//
// Discord must reach it over HTTPS: deploy it, or run a tunnel to your laptop.
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-096-discord-bot/solution"
)

var katas = []discordbot.Kata{
	{Day: 93, Name: "Time Travel Store", Prompt: "SET and DEL change a key at a time; GET asks for the value at a time.",
		Question: "SET float 4500 10; SET float 3200 30; GET float 15", Answer: "4500",
		Hints: []string{"Keep every version of each key, in time order.", "Binary search for the last version at or before t."}},
	{Day: 94, Name: "Search Suggestions", Prompt: "Print the first three titles, alphabetically, that start with the prefix.",
		Question: "titles: developer designer devops data driver dentist; prefix: de", Answer: "dentist designer developer",
		Hints: []string{"Sort the titles once.", "sort.SearchStrings finds where the prefix would go."}},
	{Day: 95, Name: "Replay Window", Prompt: "W=300. Is each event ACCEPT or DUPLICATE?",
		Question: "100 evt_a, 250 evt_a, 400 evt_a", Answer: "ACCEPT DUPLICATE ACCEPT",
		Hints: []string{"Remember when each ID was last accepted.", "At exactly W seconds, the window is over."}},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "katabot:", err)
		os.Exit(1)
	}
}

func run() error {
	keyHex, ok := os.LookupEnv("DISCORD_PUBLIC_KEY")
	if !ok {
		return fmt.Errorf("DISCORD_PUBLIC_KEY is not set (copy it from your app's General Information page)")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return fmt.Errorf("DISCORD_PUBLIC_KEY must be %d bytes of hex", ed25519.PublicKeySize)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bot := discordbot.NewBot(katas, discordbot.NewMemoryProgress(), time.Now)
	mux := http.NewServeMux()
	mux.Handle("/interactions", discordbot.NewInteractionsHandler(ed25519.PublicKey(key), bot))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	return srv.ListenAndServe()
}
