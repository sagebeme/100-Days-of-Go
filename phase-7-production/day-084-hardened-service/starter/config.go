package hardenedservice

import (
	"os"
	"strconv"
	"time"
)

// Secret is meant to hold values that must never be printed or logged.
// Right now it's just a string.
type Secret string

// Config is everything the service reads from its environment.
type Config struct {
	Addr              string        // TWENDE_ADDR, default "localhost:8080"
	Env               string        // TWENDE_ENV: "dev" (default) or "prod"
	APIKey            Secret        // TWENDE_API_KEY: required, at least 32 characters
	RatePerSecond     float64       // TWENDE_RATE: requests a second per client IP, default 5, above 0
	Burst             int           // TWENDE_BURST: default 10, at least 1
	HandlerTimeout    time.Duration // TWENDE_HANDLER_TIMEOUT: default 2s, above 0
	ShutdownGrace     time.Duration // TWENDE_SHUTDOWN_GRACE: default 10s, above 0
	ReadHeaderTimeout time.Duration // fixed: 5s
	ReadTimeout       time.Duration // fixed: 10s
	WriteTimeout      time.Duration // fixed: 10s
	IdleTimeout       time.Duration // fixed: 60s
	MaxBodyBytes      int64         // fixed: 1 MiB
}

// LoadConfig reads the TWENDE_ variables through getenv. It doesn't check anything
// yet: a typo in a number silently becomes 0.
func LoadConfig(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	cfg := Config{Addr: "localhost:8080", Env: "dev", RatePerSecond: 5, Burst: 10, HandlerTimeout: 2 * time.Second, ShutdownGrace: 10 * time.Second}
	if v := getenv("TWENDE_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := getenv("TWENDE_ENV"); v != "" {
		cfg.Env = v
	}
	cfg.APIKey = Secret(getenv("TWENDE_API_KEY"))
	if v := getenv("TWENDE_RATE"); v != "" {
		cfg.RatePerSecond, _ = strconv.ParseFloat(v, 64)
	}
	if v := getenv("TWENDE_BURST"); v != "" {
		cfg.Burst, _ = strconv.Atoi(v)
	}
	if v := getenv("TWENDE_HANDLER_TIMEOUT"); v != "" {
		cfg.HandlerTimeout, _ = time.ParseDuration(v)
	}
	if v := getenv("TWENDE_SHUTDOWN_GRACE"); v != "" {
		cfg.ShutdownGrace, _ = time.ParseDuration(v)
	}
	return cfg, nil
}
