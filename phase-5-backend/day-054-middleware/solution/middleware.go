// Package middleware wraps the matatu-booking API's handlers with the things every
// request needs: a request ID, one structured log line, and recovery from panics.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// Middleware wraps a handler in another handler.
type Middleware func(http.Handler) http.Handler

// Chain wraps h in mws so that the first middleware is the outermost: it sees the
// request first and the response last.
//
//	Chain(h, a, b)  is  a(b(h))
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ctxKey is unexported, so no other package can read or overwrite our value by
// accident with a key that happens to be equal.
type ctxKey struct{}

// RequestIDFrom returns the request ID that RequestID stored in ctx, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// RequestID gives every request an ID. It keeps an incoming X-Request-ID header if
// it is 1 to 64 characters from A-Z, a-z, 0-9, '-', '_' and '.', so an ID can follow
// a request through several services. Otherwise it makes a new one: 32 hex digits
// from crypto/rand. The ID goes in the request's context and in the X-Request-ID
// response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !validID(id) {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func validID(id string) bool {
	if len(id) < 1 || len(id) > 64 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.'
		if !ok {
			return false
		}
	}
	return true
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b) // since Go 1.24 this never returns an error
	return hex.EncodeToString(b)
}

// recorder remembers the status and size of what a handler wrote.
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rec *recorder) WriteHeader(status int) {
	if rec.status == 0 {
		rec.status = status
	}
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *recorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK // the first Write sends 200 if WriteHeader wasn't called
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the real ResponseWriter, for flushing
// and deadlines.
func (rec *recorder) Unwrap() http.ResponseWriter { return rec.ResponseWriter }

// Logger writes one log line per request, after it finishes, with the message
// "request" and these attributes:
//
//	method, path (no query string: it can hold tokens), status, bytes,
//	duration, and request_id when RequestID ran first
//
// 5xx answers log at Error level, everything else at Info. Headers are never
// logged, so Authorization and cookies stay out of the logs.
func Logger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &recorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK // the handler wrote nothing at all
			}

			level := slog.LevelInfo
			if rec.status >= 500 {
				level = slog.LevelError
			}
			logger.LogAttrs(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", RequestIDFrom(r.Context())),
			)
		})
	}
}

// Recover turns a panic in a handler into a 500 with {"error": "something went
// wrong"}, and logs the panic value and stack trace at Error level with the message
// "panic" and the request_id. One bad request then can't take the server down.
// http.ErrAbortHandler is re-panicked: it's how a handler asks the server to drop
// the connection on purpose.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				p := recover()
				if p == nil {
					return
				}
				if p == http.ErrAbortHandler {
					panic(p)
				}
				logger.ErrorContext(r.Context(), "panic",
					slog.Any("value", p),
					slog.String("stack", string(debug.Stack())),
					slog.String("request_id", RequestIDFrom(r.Context())),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "something went wrong"})
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// MaskPhone hides the middle of a phone number for logs: the first 4 and last 3
// characters stay, the rest become '*'. Anything shorter than 8 characters is all
// stars, so a short value never leaks whole.
//
//	MaskPhone("254712345678")  →  "2547*****678"
//	MaskPhone("0712")          →  "****"
func MaskPhone(phone string) string {
	if len(phone) < 8 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:4] + strings.Repeat("*", len(phone)-7) + phone[len(phone)-3:]
}
