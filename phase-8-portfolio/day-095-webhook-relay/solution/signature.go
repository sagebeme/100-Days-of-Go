// Package webhookrelay receives signed webhooks, checks them, drops repeats, and
// forwards each new event to every subscriber, retrying with exponential backoff.
package webhookrelay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
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

// mac returns HMAC-SHA256(secret, "<t>.<body>"). Signing the timestamp with the body
// means an attacker can't put a fresh time on an old request.
func mac(secret []byte, t int64, body []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(strconv.FormatInt(t, 10)))
	h.Write([]byte("."))
	h.Write(body)
	return h.Sum(nil)
}

// Sign returns the header value for body, signed with secret at time t.
func Sign(secret []byte, t time.Time, body []byte) string {
	ts := t.Unix()
	return "t=" + strconv.FormatInt(ts, 10) + ",v1=" + hex.EncodeToString(mac(secret, ts, body))
}

// Verify checks header against body. It checks the format first, then the signature
// with a constant-time compare, then that the timestamp is within tolerance of now.
func Verify(secret []byte, header string, body []byte, now time.Time, tolerance time.Duration) error {
	var ts int64
	var haveT bool
	var sigs [][]byte
	for part := range strings.SplitSeq(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return ErrMalformedSignature
		}
		switch k {
		case "t":
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil || haveT {
				return ErrMalformedSignature
			}
			ts, haveT = n, true
		case "v1":
			b, err := hex.DecodeString(v)
			if err != nil || len(b) != sha256.Size {
				return ErrMalformedSignature
			}
			sigs = append(sigs, b)
		}
		// Unknown keys (a future v2) are ignored.
	}
	if !haveT || len(sigs) == 0 {
		return ErrMalformedSignature
	}
	want := mac(secret, ts, body)
	match := false
	for _, s := range sigs {
		if hmac.Equal(s, want) {
			match = true
		}
	}
	if !match {
		return ErrBadSignature
	}
	age := now.Sub(time.Unix(ts, 0))
	if age > tolerance || age < -tolerance {
		return ErrStaleTimestamp
	}
	return nil
}
