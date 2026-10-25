package webhookrelay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	secret = []byte("whsec_test_kopa_pesa_2026")
	t0     = time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	body   = []byte(`{"id":"evt_001","type":"payment.received","amount_ksh":1500}`)
)

func TestSignMatchesSpec(t *testing.T) {
	// The spec: HMAC-SHA256 over "<unix seconds>.<body>", hex-encoded.
	h := hmac.New(sha256.New, secret)
	h.Write([]byte("1792832400."))
	h.Write(body)
	want := "t=1792832400,v1=" + hex.EncodeToString(h.Sum(nil))
	if got := Sign(secret, t0, body); got != want {
		t.Errorf("Sign\n got: %s\nwant: %s", got, want)
	}
}

func TestVerify(t *testing.T) {
	good := Sign(secret, t0, body)
	other := Sign([]byte("old-secret"), t0, body)
	sig := strings.TrimPrefix(good, "t=1792832400,")
	tests := []struct {
		name   string
		header string
		body   []byte
		now    time.Time
		want   error
	}{
		{"good", good, body, t0, nil},
		{"4 minutes late is fine", good, body, t0.Add(4 * time.Minute), nil},
		{"exactly at the tolerance is fine", good, body, t0.Add(5 * time.Minute), nil},
		{"clock 2 minutes behind is fine", good, body, t0.Add(-2 * time.Minute), nil},
		{"spaces after commas", strings.ReplaceAll(good, ",", ", "), body, t0, nil},
		{"two v1s while rotating secrets", other + "," + sig, body, t0, nil},
		{"unknown v2 is ignored", good + ",v2=abc", body, t0, nil},
		{"replayed 6 minutes later", good, body, t0.Add(6 * time.Minute), ErrStaleTimestamp},
		{"from the future", good, body, t0.Add(-6 * time.Minute), ErrStaleTimestamp},
		{"body changed", good, []byte(`{"id":"evt_001","type":"payment.received","amount_ksh":150000}`), t0, ErrBadSignature},
		{"wrong secret", other, body, t0, ErrBadSignature},
		{"fresh time on an old signature", "t=1792832700," + sig, body, t0.Add(5 * time.Minute), ErrBadSignature},
		{"empty header", "", body, t0, ErrMalformedSignature},
		{"no timestamp", sig, body, t0, ErrMalformedSignature},
		{"no v1", "t=1792832400", body, t0, ErrMalformedSignature},
		{"timestamp not a number", "t=yesterday," + sig, body, t0, ErrMalformedSignature},
		{"v1 not hex", "t=1792832400,v1=zz", body, t0, ErrMalformedSignature},
		{"v1 too short", "t=1792832400,v1=abcd", body, t0, ErrMalformedSignature},
		{"no equals sign", "t1792832400," + sig, body, t0, ErrMalformedSignature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(secret, tt.header, tt.body, tt.now, 5*time.Minute)
			if tt.want == nil && err != nil {
				t.Fatalf("Verify = %v, want nil", err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("Verify = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second},  // capped
		{60, 30 * time.Second}, // no overflow
	}
	for _, tt := range tests {
		if got := Backoff(tt.attempt, time.Second, 30*time.Second); got != tt.want {
			t.Errorf("Backoff(%d, 1s, 30s) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestRetryable(t *testing.T) {
	for status, want := range map[int]bool{200: false, 400: false, 401: false, 404: false, 408: true, 410: false, 429: true, 500: true, 502: true, 503: true} {
		if got := Retryable(status); got != want {
			t.Errorf("Retryable(%d) = %v, want %v", status, got, want)
		}
	}
}

func TestDeduper(t *testing.T) {
	d := NewDeduper(time.Hour)
	steps := []struct {
		id    string
		after time.Duration
		want  bool
	}{
		{"evt_1", 0, false},
		{"evt_1", time.Minute, true},
		{"evt_2", 2 * time.Minute, false},
		{"evt_1", 59 * time.Minute, true},
		{"evt_1", 60 * time.Minute, false}, // an hour later, evt_1 is forgotten and remembered again
		{"evt_2", 61 * time.Minute, true},
		{"evt_2", 62 * time.Minute, false},
	}
	for _, s := range steps {
		if got := d.Seen(s.id, t0.Add(s.after)); got != s.want {
			t.Errorf("at +%v, Seen(%q) = %v, want %v", s.after, s.id, got, s.want)
		}
	}
	if got := d.Len(t0.Add(62 * time.Minute)); got != 2 {
		t.Errorf("Len = %d, want 2", got)
	}
	if got := d.Len(t0.Add(5 * time.Hour)); got != 0 {
		t.Errorf("Len five hours later = %d, want 0: old IDs must be pruned", got)
	}
	d.Seen("evt_3", t0.Add(5*time.Hour))
	d.Forget("evt_3")
	if d.Seen("evt_3", t0.Add(5*time.Hour)) {
		t.Error("Seen after Forget = true, want false")
	}
}
