// Package middleware wraps the matatu-booking API's handlers with the things every
// request needs: a request ID, one structured log line, and recovery from panics.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

// Middleware wraps a handler in another handler.
type Middleware func(http.Handler) http.Handler

// Chain wraps h in mws so that the first middleware is the outermost: it sees the
// request first and the response last.
//
//	Chain(h, a, b)  is  a(b(h))
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	// Your code here. Follow the steps in the README.
	return h
}

// ctxKey is unexported, so no other package can read or overwrite our value by
// accident with a key that happens to be equal.
type ctxKey struct{}

// RequestIDFrom returns the request ID that RequestID stored in ctx, or "".
func RequestIDFrom(ctx context.Context) string {
	// Your code here.
	return ""
}

// RequestID gives every request an ID. It keeps an incoming X-Request-ID header if
// it is 1 to 64 characters from A-Z, a-z, 0-9, '-', '_' and '.', so an ID can follow
// a request through several services. Otherwise it makes a new one: 32 hex digits
// from crypto/rand. The ID goes in the request's context and in the X-Request-ID
// response header.
func RequestID(next http.Handler) http.Handler {
	// Your code here.
	return next
}

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
		// Your code here.
		return next
	}
}

// Recover turns a panic in a handler into a 500 with {"error": "something went
// wrong"}, and logs the panic value and stack trace at Error level with the message
// "panic" and the request_id. One bad request then can't take the server down.
// http.ErrAbortHandler is re-panicked: it's how a handler asks the server to drop
// the connection on purpose.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		// Your code here.
		return next
	}
}

// MaskPhone hides the middle of a phone number for logs: the first 4 and last 3
// characters stay, the rest become '*'. Anything shorter than 8 characters is all
// stars, so a short value never leaks whole.
//
//	MaskPhone("254712345678")  →  "2547*****678"
//	MaskPhone("0712")          →  "****"
func MaskPhone(phone string) string {
	// Your code here.
	return phone
}
