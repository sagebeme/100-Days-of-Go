// Package sessionsandjwts logs Twende passengers in two ways: a cookie session for
// the website, and short-lived JWT access tokens with rotating refresh tokens for
// the phone app.
package sessionsandjwts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Errors returned by this package.
var (
	ErrInvalidToken   = errors.New("invalid access token")
	ErrInvalidRefresh = errors.New("invalid refresh token")
	ErrRefreshReused  = errors.New("refresh token reused: every token in its family is revoked")
)

// Leeway is how much clock difference between servers the token checks forgive.
const Leeway = 30 * time.Second

// Config holds the token settings, read from env variables.
type Config struct {
	JWTSecret  []byte
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func durationEnv(key string, def time.Duration) (time.Duration, error) {
	s, ok := os.LookupEnv(key)
	if !ok || s == "" {
		return def, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 10s or 2m, got %q", key, s)
	}
	return d, nil
}

func stringEnv(key, def string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return def
}

// LoadConfig reads JWT_SECRET (required, at least 32 bytes), JWT_ISSUER (default
// "twende-api"), JWT_AUDIENCE (default "twende-app"), ACCESS_TOKEN_TTL (default 15m,
// between 1m and 1h) and REFRESH_TOKEN_TTL (default 720h). It reports every problem at once.
func LoadConfig() (Config, error) {
	var cfg Config
	var errs []error

	cfg.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET is required and must be at least 32 bytes"))
	}
	cfg.Issuer = stringEnv("JWT_ISSUER", "twende-api")
	cfg.Audience = stringEnv("JWT_AUDIENCE", "twende-app")

	var err error
	if cfg.AccessTTL, err = durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		errs = append(errs, err)
	} else if cfg.AccessTTL < time.Minute || cfg.AccessTTL > time.Hour {
		errs = append(errs, fmt.Errorf("ACCESS_TOKEN_TTL must be between 1m and 1h, got %s", cfg.AccessTTL))
	}
	if cfg.RefreshTTL, err = durationEnv("REFRESH_TOKEN_TTL", 720*time.Hour); err != nil {
		errs = append(errs, err)
	} else if cfg.RefreshTTL <= 0 {
		errs = append(errs, fmt.Errorf("REFRESH_TOKEN_TTL must be positive, got %s", cfg.RefreshTTL))
	}
	return cfg, errors.Join(errs...)
}

// NewSessionID returns a random, unguessable session ID from crypto/rand.
func NewSessionID() string {
	return rand.Text()
}

// NewSessionCookie returns the cookie that carries a session ID: HttpOnly so
// JavaScript can't read it, Secure so it only travels over HTTPS, SameSite=Lax so
// other sites can't send it with their POSTs, Path=/ and the given lifetime.
func NewSessionCookie(name, value string, maxAge time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge / time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

// Claims are the fields inside a Twende access token.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Tokens signs and checks access tokens.
type Tokens struct {
	Secret   []byte
	Issuer   string
	Audience string
	TTL      time.Duration
	Now      func() time.Time // time.Now in production; a fixed clock in tests
}

// NewAccessToken returns an HS256 token for userID with role, carrying sub, iss, aud,
// iat, nbf, exp (iat plus TTL) and a random jti.
func (t *Tokens) NewAccessToken(userID, role string) (string, error) {
	now := t.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    t.Issuer,
			Audience:  jwt.ClaimStrings{t.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.TTL)),
			ID:        rand.Text(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.Secret)
}

// Parse checks a token's signature and claims and returns the claims. Only HS256 is
// accepted, exp is required, iss and aud must match, iat may not be in the future, and
// Leeway is allowed for clock skew. Any failure is wrapped in ErrInvalidToken.
func (t *Tokens) Parse(token string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims,
		func(*jwt.Token) (any, error) { return t.Secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(t.Issuer),
		jwt.WithAudience(t.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(Leeway),
		jwt.WithTimeFunc(t.Now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: no subject", ErrInvalidToken)
	}
	return &claims, nil
}

type refreshRecord struct {
	userID  string
	family  string
	expires time.Time
	used    bool
	revoked bool
}

// RefreshStore keeps refresh tokens in memory, safe for concurrent use. It stores the
// SHA-256 of each token, never the token itself.
type RefreshStore struct {
	TTL time.Duration
	Now func() time.Time

	mu     sync.Mutex
	tokens map[string]*refreshRecord // keyed by hashToken(token)
}

// NewRefreshStore returns an empty store whose tokens last ttl.
func NewRefreshStore(ttl time.Duration, now func() time.Time) *RefreshStore {
	return &RefreshStore{TTL: ttl, Now: now, tokens: map[string]*refreshRecord{}}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// add saves a new token in family and returns it. The caller holds s.mu.
func (s *RefreshStore) add(userID, family string) string {
	token := rand.Text()
	s.tokens[hashToken(token)] = &refreshRecord{
		userID:  userID,
		family:  family,
		expires: s.Now().Add(s.TTL),
	}
	return token
}

// Issue starts a new token family for userID, at login, and returns its first token.
func (s *RefreshStore) Issue(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.add(userID, rand.Text())
}

// Rotate trades a refresh token for a new one in the same family and returns it with
// the user ID. Unknown, expired or revoked tokens give ErrInvalidRefresh. A token that
// was already used gives ErrRefreshReused, and the whole family is revoked, because
// either the passenger or a thief has a stolen copy.
func (s *RefreshStore) Rotate(token string) (newToken, userID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.tokens[hashToken(token)]
	if !ok || rec.revoked || !s.Now().Before(rec.expires) {
		return "", "", ErrInvalidRefresh
	}
	if rec.used {
		for _, r := range s.tokens {
			if r.family == rec.family {
				r.revoked = true
			}
		}
		return "", "", ErrRefreshReused
	}
	rec.used = true
	return s.add(rec.userID, rec.family), rec.userID, nil
}

type ctxKey struct{}

// ClaimsFrom returns the claims RequireAuth stored in ctx.
func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok
}

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// RequireAuth reads "Authorization: Bearer <token>", checks it with t and passes the
// claims to next in the request context. With no token it answers 401 with
// "WWW-Authenticate: Bearer"; with a bad one, 401 with
// `WWW-Authenticate: Bearer error="invalid_token"`.
func RequireAuth(t *Tokens, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || raw == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "log in first")
			return
		}
		claims, err := t.Parse(raw)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, http.StatusUnauthorized, "token is invalid or expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims)))
	})
}
