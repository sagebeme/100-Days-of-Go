package matatubookingapi

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
	// Your code here
	return Config{}, nil
}

// NormalizePhone accepts a Kenyan mobile number as 07XXXXXXXX, 01XXXXXXXX,
// 2547XXXXXXXX or +2547XXXXXXXX (spaces allowed) and returns the 10-digit 0 form.
func NormalizePhone(s string) (string, bool) {
	// Your code here
	return "", false
}
