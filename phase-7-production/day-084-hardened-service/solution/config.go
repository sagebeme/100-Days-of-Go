package hardenedservice

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// Secret is a string that prints and logs as "[REDACTED]" (Day 83).
type Secret string

// String hides the value from fmt's %s and %v.
func (s Secret) String() string {
	return "[REDACTED]"
}

// GoString hides the value from fmt's %#v.
func (s Secret) GoString() string {
	return "[REDACTED]"
}

// LogValue hides the value from slog.
func (s Secret) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

// MarshalJSON hides the value from encoding/json, which ignores String and is what
// slog's JSON handler uses for structs.
func (s Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

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

// LogValue logs the config as a group of fields with the API key hidden, so
// log.Info("config", "config", cfg) is safe with any handler.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("addr", c.Addr),
		slog.String("env", c.Env),
		slog.Any("api_key", c.APIKey),
		slog.Float64("rate_per_second", c.RatePerSecond),
		slog.Int("burst", c.Burst),
		slog.Duration("handler_timeout", c.HandlerTimeout),
		slog.Duration("shutdown_grace", c.ShutdownGrace),
	)
}

// LoadConfig reads the TWENDE_ variables through getenv (os.Getenv in production) and
// checks every one. It reports all the problems at once with errors.Join, each naming
// its variable, so one failed deploy shows everything to fix. It never puts a secret's
// value in an error.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:              "localhost:8080",
		Env:               "dev",
		RatePerSecond:     5,
		Burst:             10,
		HandlerTimeout:    2 * time.Second,
		ShutdownGrace:     10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxBodyBytes:      1 << 20,
	}
	var errs []error
	bad := func(name, format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s: "+format, append([]any{name}, args...)...))
	}

	if v := getenv("TWENDE_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := getenv("TWENDE_ENV"); v != "" {
		cfg.Env = v
	}
	if cfg.Env != "dev" && cfg.Env != "prod" {
		bad("TWENDE_ENV", "must be dev or prod, got %q", cfg.Env)
	}

	key := getenv("TWENDE_API_KEY")
	switch {
	case key == "":
		bad("TWENDE_API_KEY", "is required")
	case len(key) < 32:
		bad("TWENDE_API_KEY", "must be at least 32 characters, got %d", len(key))
	}
	cfg.APIKey = Secret(key)

	if v := getenv("TWENDE_RATE"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			bad("TWENDE_RATE", "must be a number above 0, got %q", v)
		}
		cfg.RatePerSecond = f
	}
	if v := getenv("TWENDE_BURST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			bad("TWENDE_BURST", "must be a whole number of at least 1, got %q", v)
		}
		cfg.Burst = n
	}
	for _, d := range []struct {
		name string
		dst  *time.Duration
	}{
		{"TWENDE_HANDLER_TIMEOUT", &cfg.HandlerTimeout},
		{"TWENDE_SHUTDOWN_GRACE", &cfg.ShutdownGrace},
	} {
		if v := getenv(d.name); v != "" {
			t, err := time.ParseDuration(v)
			if err != nil || t <= 0 {
				bad(d.name, "must be a duration above 0 such as 2s, got %q", v)
			}
			*d.dst = t
		}
	}
	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
