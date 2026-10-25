package matatubookingapi

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config is every setting the API needs.
type Config struct {
	Port            int
	JWTSecret       []byte
	Issuer          string
	AccessTTL       time.Duration
	AllowedOrigins  []string
	RatePerSecond   float64 // requests per second per IP, after the burst
	RateBurst       int
	BcryptCost      int
	AdminPhones     []string // these phones get the admin role when they register
	ClientIPHeader  string   // the header the platform's proxy puts the client IP in; "" means RemoteAddr
	ShutdownTimeout time.Duration

	Now func() time.Time // time.Now unless a test sets a fake clock
}

func intEnv(key string, def int) (int, error) {
	s := os.Getenv(key)
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q", key, s)
	}
	return n, nil
}

func floatEnv(key string, def float64) (float64, error) {
	s := os.Getenv(key)
	if s == "" {
		return def, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number, got %q", key, s)
	}
	return f, nil
}

func durationEnv(key string, def time.Duration) (time.Duration, error) {
	s := os.Getenv(key)
	if s == "" {
		return def, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 10s or 2m, got %q", key, s)
	}
	return d, nil
}

func listEnv(key string) []string {
	var out []string
	for _, s := range strings.Split(os.Getenv(key), ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// LoadConfig reads the environment and reports every problem at once:
//
//	PORT                  default 8080, 1 to 65535
//	JWT_SECRET            required, at least 32 bytes
//	JWT_ISSUER            default "matatu-api"
//	ACCESS_TOKEN_TTL      default 15m, 1m to 1h
//	CORS_ALLOWED_ORIGINS  comma-separated origins like https://app.example; default none
//	RATE_LIMIT_RPS        default 5, more than 0
//	RATE_LIMIT_BURST      default 20, at least 1
//	BCRYPT_COST           default 12, bcrypt.MinCost to bcrypt.MaxCost
//	ADMIN_PHONES          comma-separated, normalised with NormalizePhone
//	CLIENT_IP_HEADER      e.g. Fly-Client-IP or X-Forwarded-For; only behind a proxy that sets it
//	SHUTDOWN_TIMEOUT      default 10s
func LoadConfig() (Config, error) {
	cfg := Config{Now: time.Now}
	var errs []error
	var err error

	if cfg.Port, err = intEnv("PORT", 8080); err != nil {
		errs = append(errs, err)
	} else if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port))
	}
	cfg.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET is required and must be at least 32 bytes"))
	}
	cfg.Issuer = os.Getenv("JWT_ISSUER")
	if cfg.Issuer == "" {
		cfg.Issuer = "matatu-api"
	}
	if cfg.AccessTTL, err = durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		errs = append(errs, err)
	} else if cfg.AccessTTL < time.Minute || cfg.AccessTTL > time.Hour {
		errs = append(errs, fmt.Errorf("ACCESS_TOKEN_TTL must be between 1m and 1h, got %s", cfg.AccessTTL))
	}
	for _, o := range listEnv("CORS_ALLOWED_ORIGINS") {
		u, err := url.Parse(o)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.Path != "" || u.RawQuery != "" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not an origin like https://app.example", o))
			continue
		}
		cfg.AllowedOrigins = append(cfg.AllowedOrigins, strings.ToLower(o))
	}
	if cfg.RatePerSecond, err = floatEnv("RATE_LIMIT_RPS", 5); err != nil {
		errs = append(errs, err)
	} else if cfg.RatePerSecond <= 0 {
		errs = append(errs, errors.New("RATE_LIMIT_RPS must be more than 0"))
	}
	if cfg.RateBurst, err = intEnv("RATE_LIMIT_BURST", 20); err != nil {
		errs = append(errs, err)
	} else if cfg.RateBurst < 1 {
		errs = append(errs, errors.New("RATE_LIMIT_BURST must be at least 1"))
	}
	if cfg.BcryptCost, err = intEnv("BCRYPT_COST", 12); err != nil {
		errs = append(errs, err)
	} else if cfg.BcryptCost < bcrypt.MinCost || cfg.BcryptCost > bcrypt.MaxCost {
		errs = append(errs, fmt.Errorf("BCRYPT_COST must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost))
	}
	for _, p := range listEnv("ADMIN_PHONES") {
		n, ok := NormalizePhone(p)
		if !ok {
			errs = append(errs, fmt.Errorf("ADMIN_PHONES: %q is not a Kenyan mobile number", p))
			continue
		}
		cfg.AdminPhones = append(cfg.AdminPhones, n)
	}
	cfg.ClientIPHeader = os.Getenv("CLIENT_IP_HEADER")
	if cfg.ShutdownTimeout, err = durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		errs = append(errs, err)
	}
	return cfg, errors.Join(errs...)
}

// NormalizePhone accepts a Kenyan mobile number as 07XXXXXXXX, 01XXXXXXXX,
// 2547XXXXXXXX or +2547XXXXXXXX (spaces allowed) and returns the 10-digit 0 form.
func NormalizePhone(s string) (string, bool) {
	s = strings.ReplaceAll(s, " ", "")
	if rest, ok := strings.CutPrefix(s, "+254"); ok {
		s = "0" + rest
	} else if rest, ok := strings.CutPrefix(s, "254"); ok && len(s) == 12 {
		s = "0" + rest
	}
	if len(s) != 10 || (!strings.HasPrefix(s, "07") && !strings.HasPrefix(s, "01")) {
		return "", false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	return s, true
}
