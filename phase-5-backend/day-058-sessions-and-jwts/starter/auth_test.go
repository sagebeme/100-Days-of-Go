package sessionsandjwts

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secret = []byte("test-only-secret-that-is-32-bytes!!")
	start  = time.Date(2026, 3, 2, 6, 30, 0, 0, time.UTC)
)

// clock is a fake time.Now the tests move forward by hand.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Add(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.t = c.t.Add(d) }

func newTokens(c *clock) *Tokens {
	return &Tokens{Secret: secret, Issuer: "twende-api", Audience: "twende-app", TTL: 15 * time.Minute, Now: c.Now}
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", string(secret))
	t.Setenv("JWT_ISSUER", "")
	t.Setenv("JWT_AUDIENCE", "")
	t.Setenv("ACCESS_TOKEN_TTL", "")
	t.Setenv("REFRESH_TOKEN_TTL", "")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if string(cfg.JWTSecret) != string(secret) || cfg.Issuer != "twende-api" || cfg.Audience != "twende-app" ||
		cfg.AccessTTL != 15*time.Minute || cfg.RefreshTTL != 720*time.Hour {
		t.Errorf("LoadConfig() with defaults\n got: %+v\nwant: secret set, twende-api, twende-app, 15m, 720h", cfg)
	}

	t.Setenv("ACCESS_TOKEN_TTL", "5m")
	t.Setenv("JWT_ISSUER", "twende-staging")
	cfg, err = LoadConfig()
	if err != nil || cfg.AccessTTL != 5*time.Minute || cfg.Issuer != "twende-staging" {
		t.Errorf("LoadConfig() with ACCESS_TOKEN_TTL=5m JWT_ISSUER=twende-staging\n got: %+v, %v", cfg, err)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name, secret, access, refresh string
		wantInErr                     []string
	}{
		{"missing secret", "", "", "", []string{"JWT_SECRET"}},
		{"short secret", "only-twenty-bytes!!!", "", "", []string{"JWT_SECRET"}},
		{"bad ttl", string(secret), "15", "", []string{"ACCESS_TOKEN_TTL"}},
		{"ttl too long", string(secret), "24h", "", []string{"ACCESS_TOKEN_TTL"}},
		{"every problem at once", "", "soon", "forever", []string{"JWT_SECRET", "ACCESS_TOKEN_TTL", "REFRESH_TOKEN_TTL"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.secret)
			t.Setenv("ACCESS_TOKEN_TTL", tt.access)
			t.Setenv("REFRESH_TOKEN_TTL", tt.refresh)
			_, err := LoadConfig()
			if err == nil {
				t.Fatalf("LoadConfig() error = nil, want one naming %v", tt.wantInErr)
			}
			for _, w := range tt.wantInErr {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("LoadConfig() error = %q, want it to mention %s", err, w)
				}
			}
			if tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
				t.Errorf("the error message contains the secret itself: %q", err)
			}
		})
	}
}

func TestSessionCookie(t *testing.T) {
	id := NewSessionID()
	if len(id) < 26 || id == NewSessionID() {
		t.Fatalf("NewSessionID() = %q: want at least 26 random characters, different every call", id)
	}
	c := NewSessionCookie("twende_session", id, 2*time.Hour)
	if c.Name != "twende_session" || c.Value != id || c.Path != "/" || c.MaxAge != 7200 {
		t.Errorf("cookie = %+v\nwant Name twende_session, the ID as Value, Path /, MaxAge 7200", c)
	}
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie flags: HttpOnly=%v Secure=%v SameSite=%v; want true, true, Lax", c.HttpOnly, c.Secure, c.SameSite)
	}
	line := c.String()
	for _, want := range []string{"HttpOnly", "Secure", "SameSite=Lax", "Max-Age=7200"} {
		if !strings.Contains(line, want) {
			t.Errorf("Set-Cookie %q is missing %s", line, want)
		}
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	c := &clock{t: start}
	tk := newTokens(c)
	raw, err := tk.NewAccessToken("0712345678", "passenger")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}
	if strings.Count(raw, ".") != 2 {
		t.Fatalf("NewAccessToken = %q, want header.payload.signature", raw)
	}
	claims, err := tk.Parse(raw)
	if err != nil {
		t.Fatalf("Parse(fresh token): %v", err)
	}
	if claims.Subject != "0712345678" || claims.Role != "passenger" || claims.Issuer != "twende-api" {
		t.Errorf("claims = sub %q role %q iss %q; want 0712345678, passenger, twende-api", claims.Subject, claims.Role, claims.Issuer)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(start.Add(15*time.Minute)) {
		t.Errorf("exp = %v, want %v", claims.ExpiresAt, start.Add(15*time.Minute))
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(start) {
		t.Errorf("iat = %v, want %v", claims.IssuedAt, start)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "twende-app" {
		t.Errorf("aud = %v, want [twende-app]", claims.Audience)
	}
	if claims.ID == "" {
		t.Errorf("jti is empty, want a random ID")
	}

	c.Add(15*time.Minute + 20*time.Second) // just past exp, inside the leeway
	if _, err := tk.Parse(raw); err != nil {
		t.Errorf("Parse 20s after exp = %v, want it accepted thanks to the 30s leeway", err)
	}
	c.Add(15 * time.Second) // now 35s past exp
	if _, err := tk.Parse(raw); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Parse 35s after exp = %v, want ErrInvalidToken", err)
	}
}

// sign makes a token by hand, the way an attacker or a buggy service might.
func sign(t *testing.T, method jwt.SigningMethod, key any, claims jwt.Claims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParseRejects(t *testing.T) {
	c := &clock{t: start}
	tk := newTokens(c)
	good := func() jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Subject:   "0712345678",
			Issuer:    "twende-api",
			Audience:  jwt.ClaimStrings{"twende-app"},
			IssuedAt:  jwt.NewNumericDate(start),
			ExpiresAt: jwt.NewNumericDate(start.Add(10 * time.Minute)),
		}
	}
	withClaims := func(edit func(*jwt.RegisteredClaims)) Claims {
		rc := good()
		edit(&rc)
		return Claims{Role: "admin", RegisteredClaims: rc}
	}

	valid := sign(t, jwt.SigningMethodHS256, secret, withClaims(func(*jwt.RegisteredClaims) {}))
	if _, err := tk.Parse(valid); err != nil {
		t.Fatalf("Parse(a valid hand-made token) = %v, want nil", err)
	}

	none := sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, withClaims(func(*jwt.RegisteredClaims) {}))
	tests := []struct {
		name  string
		token string
	}{
		{"alg none", none},
		{"HS512, not the pinned HS256", sign(t, jwt.SigningMethodHS512, secret, withClaims(func(*jwt.RegisteredClaims) {}))},
		{"signed with another secret", sign(t, jwt.SigningMethodHS256, []byte("some-other-secret-also-32-bytes-long"), withClaims(func(*jwt.RegisteredClaims) {}))},
		{"expired", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) {
			rc.ExpiresAt = jwt.NewNumericDate(start.Add(-time.Minute))
		}))},
		{"no exp", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) { rc.ExpiresAt = nil }))},
		{"issued in the future", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) {
			rc.IssuedAt = jwt.NewNumericDate(start.Add(5 * time.Minute))
		}))},
		{"wrong issuer", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) { rc.Issuer = "evil-api" }))},
		{"wrong audience", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) {
			rc.Audience = jwt.ClaimStrings{"admin-dashboard"}
		}))},
		{"no subject", sign(t, jwt.SigningMethodHS256, secret, withClaims(func(rc *jwt.RegisteredClaims) { rc.Subject = "" }))},
		{"tampered signature", valid[:len(valid)-4] + "AAAA"},
		{"garbage", "not.a.token"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tk.Parse(tt.token)
			if !errors.Is(err, ErrInvalidToken) || got != nil {
				t.Errorf("Parse\n got: %v, %v\nwant: nil, ErrInvalidToken", got, err)
			}
		})
	}
}

func TestRefreshRotation(t *testing.T) {
	c := &clock{t: start}
	s := NewRefreshStore(24*time.Hour, c.Now)

	first := s.Issue("0712345678")
	if len(first) < 26 {
		t.Fatalf("Issue = %q, want a random token of at least 26 characters", first)
	}
	second, user, err := s.Rotate(first)
	if err != nil || user != "0712345678" || second == "" || second == first {
		t.Fatalf("Rotate(first)\n got: %q, %q, %v\nwant: a new token, 0712345678, <nil>", second, user, err)
	}
	third, _, err := s.Rotate(second)
	if err != nil {
		t.Fatalf("Rotate(second) = %v, want nil", err)
	}

	// Someone replays the first token: theft. Everything in the family dies.
	if _, _, err := s.Rotate(first); !errors.Is(err, ErrRefreshReused) {
		t.Errorf("Rotate(first again) = %v, want ErrRefreshReused", err)
	}
	if _, _, err := s.Rotate(third); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate(third) after reuse = %v, want ErrInvalidRefresh: the whole family is revoked", err)
	}

	// Another login is a separate family and isn't affected.
	other := s.Issue("0722000000")
	if _, _, err := s.Rotate(other); err != nil {
		t.Errorf("Rotate(token from another family) = %v, want nil", err)
	}

	if _, _, err := s.Rotate("made-up-token"); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate(unknown) = %v, want ErrInvalidRefresh", err)
	}

	old := s.Issue("0733000000")
	c.Add(24 * time.Hour)
	if _, _, err := s.Rotate(old); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("Rotate(expired) = %v, want ErrInvalidRefresh", err)
	}
}

func TestRefreshStoreKeepsOnlyHashes(t *testing.T) {
	s := NewRefreshStore(time.Hour, time.Now)
	tok := s.Issue("0712345678")
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.tokens) != 1 {
		t.Fatalf("store has %d tokens, want 1", len(s.tokens))
	}
	for k := range s.tokens {
		if k == tok || strings.Contains(k, tok) {
			t.Errorf("the store's key is the raw token; store a SHA-256 hash instead")
		}
	}
}

func TestRefreshConcurrentReuse(t *testing.T) {
	s := NewRefreshStore(time.Hour, time.Now)
	tok := s.Issue("0712345678")
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := s.Rotate(tok); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Errorf("20 goroutines rotated the same token and %d succeeded, want exactly 1", wins)
	}
}

func TestRequireAuth(t *testing.T) {
	c := &clock{t: start}
	tk := newTokens(c)
	h := RequireAuth(tk, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFrom(r.Context())
		if !ok {
			t.Error("ClaimsFrom found no claims inside a handler RequireAuth let through")
			return
		}
		w.Write([]byte(claims.Subject))
	}))
	good, _ := tk.NewAccessToken("0712345678", "passenger")

	tests := []struct {
		name, header string
		wantStatus   int
		wantWWW      string
		wantBody     string
	}{
		{"valid token", "Bearer " + good, 200, "", "0712345678"},
		{"no header", "", 401, "Bearer", ""},
		{"wrong scheme", "Basic dXNlcjpwYXNz", 401, "Bearer", ""},
		{"bad token", "Bearer not.a.token", 401, `Bearer error="invalid_token"`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tt.wantWWW {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tt.wantWWW)
			}
			if tt.wantStatus == 200 {
				if rec.Body.String() != tt.wantBody {
					t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
				}
				return
			}
			var body struct{ Error string }
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
				t.Errorf("401 body = %q, want JSON like {\"error\": \"...\"}", rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}
