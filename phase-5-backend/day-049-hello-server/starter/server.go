// Package helloserver is the first version of the matatu-booking API: a web server
// that answers a few plain-text routes, has sensible timeouts, and shuts down
// without cutting off requests that are still running.
package helloserver

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Config holds the settings the server reads from the environment.
type Config struct {
	Port   int    // PORT, default 8080
	AppEnv string // APP_ENV, default "development"
}

// LoadConfig reads PORT and APP_ENV from the environment. A missing or empty value
// gets its default. PORT must be a whole number from 1 to 65535. APP_ENV must be
// "development", "staging" or "production". Every problem is reported at once.
//
//	PORT=9000 APP_ENV=production  →  Config{Port: 9000, AppEnv: "production"}, nil
//	PORT=abc                      →  Config{}, PORT must be a whole number from 1 to 65535, got "abc"
func LoadConfig() (Config, error) {
	// Your code here. Follow the steps in the README.
	return Config{}, nil
}

// Addr returns the address to listen on: ":8080" for port 8080. The empty host
// means every network interface on the machine.
func (c Config) Addr() string {
	// Your code here.
	return ""
}

// Routes returns the server's handler:
//
//	GET /             200 "Karibu! The matatu booking API is running.\n"
//	GET /health       200 "ok\n"
//	GET /hello/{name} 200 "Habari, {name}!\n"
//	anything else     404 from the mux
//
// Every answer is plain text.
func Routes() http.Handler {
	mux := http.NewServeMux()
	// Your code here: register the three routes.
	return mux
}

// NewServer returns an http.Server for h on addr with these timeouts:
//
//	ReadHeaderTimeout  5s   time to read the request line and headers
//	ReadTimeout       10s   time to read the whole request, body included
//	WriteTimeout      10s   time from the end of the headers to the end of the answer
//	IdleTimeout       60s   how long a kept-alive connection may sit unused
func NewServer(addr string, h http.Handler) *http.Server {
	// Your code here.
	return &http.Server{}
}

// Run serves srv on ln until ctx is cancelled, then shuts down gracefully: it stops
// accepting new connections and waits up to shutdownTimeout for requests that are
// still running. It returns nil after a clean shutdown, the error from Serve if the
// server fails first, or an error wrapping context.DeadlineExceeded if requests were
// still running when the timeout ran out.
func Run(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	// Your code here.
	return nil
}
