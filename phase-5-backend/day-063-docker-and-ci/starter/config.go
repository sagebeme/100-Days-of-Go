// Package dockerandci gets the Twende API ready to run in a container: config checked
// at startup, and the /healthz and /readyz endpoints a platform uses to decide when
// to send traffic and when to restart it.
package dockerandci

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every setting the API reads from its environment.
type Config struct {
	Port            int
	Env             string // "development" or "production"
	DatabaseURL     string
	JWTSecret       string
	ShutdownTimeout time.Duration
	ReadyTimeout    time.Duration
}

func intEnv(key string, def int) (int, error) {
	s, ok := os.LookupEnv(key)
	if !ok || s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q", key, s)
	}
	return n, nil
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

// Load reads and checks the config, and reports every problem at once:
//
//	PORT              1 to 65535, default 8080
//	APP_ENV           development (the default) or production
//	DATABASE_URL      a postgres:// or postgresql:// URL; required in production
//	JWT_SECRET        at least 32 bytes; required in production
//	SHUTDOWN_TIMEOUT  a positive duration, default 10s
//	READY_TIMEOUT     a positive duration, default 2s
//
// Error messages name the variable but never include a secret's value.
func Load() (Config, error) {
	// Your code here
	return Config{}, nil
}
