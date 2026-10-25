package discordbot

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A fixed key, so the tests are the same on every run. Never put a real key in code.
var (
	testKey = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	testPub = testKey.Public().(ed25519.PublicKey)
)

const ts = "1792832400"

func signReq(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/interactions", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(TimestampHeader, ts)
	r.Header.Set(SignatureHeader, hex.EncodeToString(ed25519.Sign(testKey, []byte(ts+body))))
	return r
}

func TestVerifyRequest(t *testing.T) {
	body := []byte(`{"type":1}`)
	good := hex.EncodeToString(ed25519.Sign(testKey, append([]byte(ts), body...)))
	otherKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, ed25519.SeedSize))
	tests := []struct {
		name string
		pub  ed25519.PublicKey
		sig  string
		ts   string
		body []byte
		want bool
	}{
		{"good", testPub, good, ts, body, true},
		{"upper-case hex is fine", testPub, strings.ToUpper(good), ts, body, true},
		{"body changed", testPub, good, ts, []byte(`{"type":2}`), false},
		{"timestamp changed", testPub, good, "1792832401", body, false},
		{"signed by another key", otherKey.Public().(ed25519.PublicKey), good, ts, body, false},
		{"no signature", testPub, "", ts, body, false},
		{"no timestamp", testPub, good, "", body, false},
		{"not hex", testPub, "zz" + good[2:], ts, body, false},
		{"too short", testPub, good[:64], ts, body, false},
		{"empty public key", nil, good, ts, body, false},
	}
	for _, tt := range tests {
		if got := VerifyRequest(tt.pub, tt.sig, tt.ts, tt.body); got != tt.want {
			t.Errorf("%s: VerifyRequest = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseCommand(t *testing.T) {
	inServer := `{"type":2,"data":{"name":"submit","options":[{"name":"day","type":4,"value":93},{"name":"answer","type":3,"value":"4500"}]},"member":{"user":{"id":"111"}}}`
	c, err := ParseCommand([]byte(inServer))
	if err != nil {
		t.Fatalf("ParseCommand: %v", err)
	}
	if c.Name != "submit" || c.UserID != "111" || c.Options["day"] != "93" || c.Options["answer"] != "4500" {
		t.Errorf("ParseCommand = %+v, want submit by 111 with day 93 and answer 4500", c)
	}
	inDM := `{"type":2,"data":{"name":"kata"},"user":{"id":"222"}}`
	if c, err := ParseCommand([]byte(inDM)); err != nil || c.UserID != "222" || c.Name != "kata" || c.Options == nil {
		t.Errorf("ParseCommand(DM) = %+v, %v; want kata by 222 and a non-nil Options map", c, err)
	}
	if _, err := ParseCommand([]byte(`{"type":3}`)); err == nil {
		t.Error("ParseCommand of a type 3 interaction = nil error, want an error")
	}
	if _, err := ParseCommand([]byte(`{`)); err == nil {
		t.Error("ParseCommand of broken JSON = nil error, want an error")
	}
}

// echo is a Handler that repeats the command, so these tests don't depend on Bot.
type echo struct{}

func (echo) Handle(_ context.Context, c Command) Response {
	return Response{Content: c.Name + " by " + c.UserID + " day " + c.Options["day"], Ephemeral: c.Name == "hint"}
}

type reply struct {
	Type int `json:"type"`
	Data *struct {
		Content string `json:"content"`
		Flags   int    `json:"flags"`
	} `json:"data"`
}

func TestInteractionsHandler(t *testing.T) {
	h := NewInteractionsHandler(testPub, echo{})
	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantType   int
		wantText   string
		wantFlags  int
	}{
		{"PING gets PONG", signReq(`{"type":1}`), 200, ResponsePong, "", 0},
		{"a command", signReq(`{"type":2,"data":{"name":"kata","options":[{"name":"day","value":94}]},"member":{"user":{"id":"42"}}}`),
			200, ResponseChannelMessageWithSource, "kata by 42 day 94", 0},
		{"an ephemeral reply", signReq(`{"type":2,"data":{"name":"hint","options":[{"name":"day","value":93}]},"user":{"id":"42"}}`),
			200, ResponseChannelMessageWithSource, "hint by 42 day 93", FlagEphemeral},
		{"bad signature", func() *http.Request {
			r := signReq(`{"type":1}`)
			r.Header.Set(SignatureHeader, hex.EncodeToString(make([]byte, ed25519.SignatureSize)))
			return r
		}(), http.StatusUnauthorized, 0, "", 0},
		{"no signature", httptest.NewRequest(http.MethodPost, "/interactions", strings.NewReader(`{"type":1}`)), http.StatusUnauthorized, 0, "", 0},
		{"GET", httptest.NewRequest(http.MethodGet, "/interactions", nil), http.StatusMethodNotAllowed, 0, "", 0},
		{"signed broken JSON", signReq(`{`), http.StatusBadRequest, 0, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, tt.req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus != 200 {
				return
			}
			if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var got reply
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("reply isn't JSON: %v\n%s", err, w.Body)
			}
			if got.Type != tt.wantType {
				t.Errorf("reply type %d, want %d", got.Type, tt.wantType)
			}
			if tt.wantText == "" {
				return
			}
			if got.Data == nil || got.Data.Content != tt.wantText || got.Data.Flags != tt.wantFlags {
				t.Errorf("reply data = %+v, want content %q and flags %d", got.Data, tt.wantText, tt.wantFlags)
			}
		})
	}
}

func TestEndToEndWithBot(t *testing.T) {
	h := NewInteractionsHandler(testPub, NewBot(katas(), NewMemoryProgress(), on(2026, 1, 1)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, signReq(`{"type":2,"data":{"name":"submit","options":[{"name":"day","value":93},{"name":"answer","value":"4500"}]},"member":{"user":{"id":"7"}}}`))
	var got reply
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.Data == nil || !strings.HasPrefix(got.Data.Content, "Correct!") || got.Data.Flags != FlagEphemeral {
		t.Errorf("submit through HTTP = %s, want an ephemeral Correct!", w.Body)
	}
}
