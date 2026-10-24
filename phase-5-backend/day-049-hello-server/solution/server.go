// Package helloserver is the first version of the matatu-booking API: a web server
// that answers a few plain-text routes, has sensible timeouts, and shuts down
// without cutting off requests that are still running.
package helloserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
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
	cfg := Config{Port: 8080, AppEnv: "development"}
	var errs []error

	if s := os.Getenv("PORT"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 65535 {
			errs = append(errs, fmt.Errorf("PORT must be a whole number from 1 to 65535, got %q", s))
		} else {
			cfg.Port = n
		}
	}

	if s := os.Getenv("APP_ENV"); s != "" {
		switch s {
		case "development", "staging", "production":
			cfg.AppEnv = s
		default:
			errs = append(errs, fmt.Errorf("APP_ENV must be development, staging or production, got %q", s))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Addr returns the address to listen on: ":8080" for port 8080. The empty host
// means every network interface on the machine.
func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.Port)
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
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /hello/{name}", hello)
	return mux
}

func home(w http.ResponseWriter, r *http.Request) {
	writeText(w, "Karibu! The matatu booking API is running.\n")
}

func health(w http.ResponseWriter, r *http.Request) {
	writeText(w, "ok\n")
}

func hello(w http.ResponseWriter, r *http.Request) {
	writeText(w, "Habari, "+r.PathValue("name")+"!\n")
}

// writeText sends s as a 200 plain-text answer. The Content-Type and nosniff headers
// stop a browser from guessing that a name like "<script>" is HTML.
func writeText(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprint(w, s)
}

// NewServer returns an http.Server for h on addr with these timeouts:
//
//	ReadHeaderTimeout  5s   time to read the request line and headers
//	ReadTimeout       10s   time to read the whole request, body included
//	WriteTimeout      10s   time from the end of the headers to the end of the answer
//	IdleTimeout       60s   how long a kept-alive connection may sit unused
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Run serves srv on ln until ctx is cancelled, then shuts down gracefully: it stops
// accepting new connections and waits up to shutdownTimeout for requests that are
// still running. It returns nil after a clean shutdown, the error from Serve if the
// server fails first, or an error wrapping context.DeadlineExceeded if requests were
// still running when the timeout ran out.
func Run(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	errc := make(chan error, 1)
	go func() {
		errc <- srv.Serve(ln)
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	// ctx is already cancelled, so the shutdown needs a fresh context of its own.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
