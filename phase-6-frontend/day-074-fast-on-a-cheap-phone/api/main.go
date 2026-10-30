// Command api serves Day 74's built front end with caching headers. Build first, then run, from
// the repo root:
//
//	(cd phase-6-frontend && npm run build -- day-074-fast-on-a-cheap-phone/starter)
//	go run ./phase-6-frontend/day-074-fast-on-a-cheap-phone/api
//
// Then open http://localhost:8080. Serving the page and any API from one Go server means one
// origin, so no CORS at all.
//
// Environment: ADDR (default :8080), STATIC_DIR (default
// phase-6-frontend/day-074-fast-on-a-cheap-phone/starter/dist).
package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// CacheControl says how long browsers and CDNs may keep a file.
//   - /assets/ files have a content hash in their name (index-B3kx9a.js). A new build makes a new
//     name, so the old file can be kept for a year and never re-checked.
//   - HTML must always be re-checked, or users would keep loading last week's asset names.
//     "no-cache" means "store it, but ask before using it"; Go answers that with a cheap 304.
//   - Anything else (favicon, robots.txt): an hour.
func CacheControl(urlPath string) string {
	switch {
	case strings.HasPrefix(urlPath, "/assets/"):
		return "public, max-age=31536000, immutable"
	case urlPath == "/" || strings.HasSuffix(urlPath, ".html"):
		return "no-cache"
	default:
		return "public, max-age=3600"
	}
}

func newHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir)) // also handles Last-Modified and If-Modified-Since → 304
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", CacheControl(r.URL.Path))
		files.ServeHTTP(w, r)
	})
}

func main() {
	addr := getenv("ADDR", ":8080")
	dir := getenv("STATIC_DIR", "phase-6-frontend/day-074-fast-on-a-cheap-phone/starter/dist")
	if _, err := os.Stat(dir); err != nil {
		log.Fatalf("%v: build the front end first (see the README)", err)
	}
	log.Printf("serving %s on %s", dir, addr)
	log.Fatal(http.ListenAndServe(addr, newHandler(dir)))
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
