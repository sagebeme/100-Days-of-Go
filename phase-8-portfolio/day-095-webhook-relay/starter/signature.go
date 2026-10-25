// Package webhookrelay receives signed webhooks, checks them, drops repeats, and
// forwards each new event to every subscriber, retrying with exponential backoff.
package webhookrelay

import (
	"errors"
	"time"
)

// SignatureHeader carries the signature, in the form "t=<unix seconds>,v1=<hex>".
// There may be more than one v1, for example while a secret is being rotated.
const SignatureHeader = "Webhook-Signature"

var (
	// ErrMalformedSignature means the header is missing or can't be parsed.
	ErrMalformedSignature = errors.New("webhookrelay: malformed signature header")
	// ErrBadSignature means no v1 matches the body.
	ErrBadSignature = errors.New("webhookrelay: signature doesn't match")
	// ErrStaleTimestamp means t is further from now than the tolerance, either way.
	ErrStaleTimestamp = errors.New("webhookrelay: timestamp outside tolerance")
)

// Sign returns the header value for body, signed with secret at time t:
// "t=<t.Unix()>,v1=<hex of HMAC-SHA256(secret, "<t.Unix()>.<body>")>".
func Sign(secret []byte, t time.Time, body []byte) string {
	// Your code here
	return ""
}

// Verify checks header against body. Check the format first (ErrMalformedSignature),
// then the signature with a constant-time compare (ErrBadSignature), then that the
// timestamp is within tolerance of now, either way (ErrStaleTimestamp). Ignore keys
// other than t and v1.
func Verify(secret []byte, header string, body []byte, now time.Time, tolerance time.Duration) error {
	// Your code here
	return nil
}
