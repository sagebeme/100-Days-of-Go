// Package dockerandci gets the Twende API ready to run in a container: config checked
// at startup, and the /healthz and /readyz endpoints a platform uses to decide when
// to send traffic and when to restart it.
package dockerandci

import (
	"errors"
	"fmt"
	"net/url"
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
	var cfg Config
	var errs []error
	var err error

	if cfg.Port, err = intEnv("PORT", 8080); err != nil {
		errs = append(errs, err)
	} else if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port))
	}

	cfg.Env = os.Getenv("APP_ENV")
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.Env != "development" && cfg.Env != "production" {
		errs = append(errs, fmt.Errorf("APP_ENV must be development or production, got %q", cfg.Env))
	}
	prod := cfg.Env == "production"

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		if prod {
			errs = append(errs, errors.New("DATABASE_URL is required in production"))
		}
	} else if u, err := url.Parse(cfg.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		errs = append(errs, errors.New("DATABASE_URL must look like postgres://user:password@host:5432/dbname"))
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" && prod {
		errs = append(errs, errors.New("JWT_SECRET is required in production"))
	} else if cfg.JWTSecret != "" && len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 bytes"))
	}

	if cfg.ShutdownTimeout, err = durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		errs = append(errs, err)
	} else if cfg.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("SHUTDOWN_TIMEOUT must be positive"))
	}
	if cfg.ReadyTimeout, err = durationEnv("READY_TIMEOUT", 2*time.Second); err != nil {
		errs = append(errs, err)
	} else if cfg.ReadyTimeout <= 0 {
		errs = append(errs, errors.New("READY_TIMEOUT must be positive"))
	}

	return cfg, errors.Join(errs...)
}
