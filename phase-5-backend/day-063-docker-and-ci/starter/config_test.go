package dockerandci

import (
	"strings"
	"testing"
	"time"
)

var vars = []string{"PORT", "APP_ENV", "DATABASE_URL", "JWT_SECRET", "SHUTDOWN_TIMEOUT", "READY_TIMEOUT"}

// setEnv clears every variable Load reads, then sets the ones given.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range vars {
		t.Setenv(k, kv[k])
	}
}

const secret = "a-test-secret-of-at-least-32-bytes!!"

func TestLoadDefaults(t *testing.T) {
	setEnv(t, nil)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with nothing set, in development: %v", err)
	}
	want := Config{Port: 8080, Env: "development", ShutdownTimeout: 10 * time.Second, ReadyTimeout: 2 * time.Second}
	if cfg != want {
		t.Errorf("Load()\n got: %+v\nwant: %+v", cfg, want)
	}
}

func TestLoadProduction(t *testing.T) {
	setEnv(t, map[string]string{
		"PORT": "10000", "APP_ENV": "production", "JWT_SECRET": secret,
		"DATABASE_URL":     "postgres://gopher:pw@db.internal:5432/matatu?sslmode=require",
		"SHUTDOWN_TIMEOUT": "25s", "READY_TIMEOUT": "500ms",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.Port != 10000 || cfg.Env != "production" || cfg.JWTSecret != secret ||
		cfg.ShutdownTimeout != 25*time.Second || cfg.ReadyTimeout != 500*time.Millisecond ||
		!strings.HasPrefix(cfg.DatabaseURL, "postgres://") {
		t.Errorf("Load() = %+v", cfg)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		mention []string
	}{
		{"port not a number", map[string]string{"PORT": "eighty"}, []string{"PORT"}},
		{"port too big", map[string]string{"PORT": "70000"}, []string{"PORT"}},
		{"port zero", map[string]string{"PORT": "0"}, []string{"PORT"}},
		{"unknown env", map[string]string{"APP_ENV": "staging"}, []string{"APP_ENV"}},
		{"production needs secrets", map[string]string{"APP_ENV": "production"}, []string{"DATABASE_URL", "JWT_SECRET"}},
		{"short secret", map[string]string{"JWT_SECRET": "short-secret"}, []string{"JWT_SECRET"}},
		{"not a postgres url", map[string]string{"DATABASE_URL": "mysql://root@localhost/matatu"}, []string{"DATABASE_URL"}},
		{"bad durations", map[string]string{"SHUTDOWN_TIMEOUT": "10", "READY_TIMEOUT": "-1s"}, []string{"SHUTDOWN_TIMEOUT", "READY_TIMEOUT"}},
		{"everything at once", map[string]string{"PORT": "x", "APP_ENV": "prod", "JWT_SECRET": "tiny"}, []string{"PORT", "APP_ENV", "JWT_SECRET"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)
			_, err := Load()
			if err == nil {
				t.Fatalf("Load() error = nil, want one naming %v", tt.mention)
			}
			for _, m := range tt.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q doesn't mention %s", err, m)
				}
			}
			for _, k := range []string{"JWT_SECRET", "DATABASE_URL"} {
				if v := tt.env[k]; v != "" && strings.Contains(err.Error(), v) {
					t.Errorf("error %q contains the value of %s", err, k)
				}
			}
		})
	}
}
