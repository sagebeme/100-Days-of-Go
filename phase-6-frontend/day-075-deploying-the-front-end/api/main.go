// Command api serves Day 75's built front end the way a static host does, plus a /config.json
// made from environment variables at runtime. Build first, then run, from the repo root:
//
//	(cd phase-6-frontend && npm run build -- day-075-deploying-the-front-end/starter)
//	API_URL=https://api.twende.example go run ./phase-6-frontend/day-075-deploying-the-front-end/api
//
// Environment: ADDR (default :8080), STATIC_DIR (default
// phase-6-frontend/day-075-deploying-the-front-end/starter/dist), and the runtime config:
// API_URL, WS_URL, SEAT_MAP. Unset ones are left out of /config.json.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RuntimeConfig is what /config.json returns. Only public values belong here: the browser shows
// it to anyone who asks.
func RuntimeConfig(getenv func(string) string) map[string]any {
	cfg := map[string]any{}
	if v := getenv("API_URL"); v != "" {
		cfg["apiUrl"] = v
	}
	if v := getenv("WS_URL"); v != "" {
		cfg["wsUrl"] = v
	}
	if v := getenv("SEAT_MAP"); v == "true" || v == "false" {
		cfg["seatMap"] = v == "true"
	}
	return cfg
}

// newHandler serves files from dir. A path with no file, like /bookings/K7Q2 that React draws
// in the browser, gets index.html (the "SPA fallback"). A missing /assets/ file is a real 404:
// answering it with HTML would only confuse the browser.
func newHandler(dir string, getenv func(string) string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/config.json":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-cache")
			json.NewEncoder(w).Encode(RuntimeConfig(getenv))
			return
		case strings.HasPrefix(r.URL.Path, "/assets/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			files.ServeHTTP(w, r)
			return
		}
		clean := path.Clean("/" + r.URL.Path)
		if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(clean))); err == nil && !info.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func main() {
	addr := getenv("ADDR", ":8080")
	dir := getenv("STATIC_DIR", "phase-6-frontend/day-075-deploying-the-front-end/starter/dist")
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		log.Fatalf("%v: build the front end first (see the README)", err)
	}
	log.Printf("serving %s on %s, config %v", dir, addr, RuntimeConfig(os.Getenv))
	log.Fatal(http.ListenAndServe(addr, newHandler(dir, os.Getenv)))
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
