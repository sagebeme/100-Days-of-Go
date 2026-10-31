// Command api is Day 68's route search API. Run it from the repo root:
//
//	go run ./phase-6-frontend/day-068-talking-to-go/api
//
// GET /api/routes?q=ngong lists routes whose destination contains q (any case).
// GET /api/routes?fail=1 answers 500, so you can see the front end's error state.
//
// Environment: ADDR (default :8080), ALLOWED_ORIGIN (default http://localhost:5173),
// DELAY (default 400ms, a Go duration like 2s, so the loading state is visible).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Route is one matatu route. The JSON names match the TypeScript interface.
type Route struct {
	Number   string `json:"number,omitempty"`
	Operator string `json:"operator"`
	BoardsAt string `json:"boardsAt"`
	GoesTo   string `json:"goesTo"`
	FareKsh  int    `json:"fareKsh,omitempty"`
}

// routes were checked on nishukishe.com in October 2026. Routes change, so confirm with the conductor.
var routes = []Route{
	{Number: "46K", Operator: "Citi Hoppa", BoardsAt: "Kencom / Ambassadeur", GoesTo: "Kawangware", FareKsh: 100},
	{Number: "111", Operator: "MCA Sacco", BoardsAt: "Railways", GoesTo: "Ngong", FareKsh: 100},
	{Number: "125", Operator: "City Shuttle", BoardsAt: "Railways", GoesTo: "Ongata Rongai", FareKsh: 100},
	{Number: "45", Operator: "Nicco Travellers", BoardsAt: "Ronald Ngala Street", GoesTo: "Githurai", FareKsh: 100},
	{Number: "58", Operator: "Buruburu 58 Travellers Sacco", BoardsAt: "Ambassadeur", GoesTo: "Buruburu", FareKsh: 150},
	{Operator: "Highrise Kibera Sacco", BoardsAt: "Railways", GoesTo: "Kibera"},
}

func main() {
	addr := getenv("ADDR", ":8080")
	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	delay, err := time.ParseDuration(getenv("DELAY", "400ms"))
	if err != nil {
		log.Fatalf("DELAY: %v", err)
	}
	log.Printf("listening on %s, allowing origin %s", addr, origin)
	log.Fatal(http.ListenAndServe(addr, newHandler(origin, delay)))
}

func newHandler(origin string, delay time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/routes", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
		case <-r.Context().Done(): // the browser aborted the request: stop work
			return
		}
		if r.URL.Query().Get("fail") == "1" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "the route database is down"})
			return
		}
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		found := []Route{} // not nil, so an empty result is [] and not null
		for _, rt := range routes {
			if strings.Contains(strings.ToLower(rt.GoesTo), q) {
				found = append(found, rt)
			}
		}
		writeJSON(w, http.StatusOK, found)
	})
	return cors(origin, mux)
}

// cors lets pages from one origin read our responses. Without it the browser still sends
// the request, but refuses to show the response to the page's JavaScript.
func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if r.Header.Get("Origin") == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
