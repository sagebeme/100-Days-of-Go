package owasptop10

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

var cfg = Config{
	Port:        8080,
	DatabaseURL: "postgres://gopher:s3cr3t-db-pass@db.internal:5432/matatu?sslmode=require",
	JWTSecret:   "jwt-secret-that-must-never-leak-123",
	MpesaKey:    "mpesa-consumer-key-ABC",
}

var leaks = []string{"s3cr3t-db-pass", "jwt-secret-that-must-never-leak-123", "mpesa-consumer-key-ABC"}

func checkNoLeak(t *testing.T, where, out string) {
	t.Helper()
	for _, s := range leaks {
		if strings.Contains(out, s) {
			t.Errorf("%s leaks %q:\n%s", where, s, out)
		}
	}
}

func TestSecretNeverPrints(t *testing.T) {
	s := Secret("hunter2-mpesa-pin")
	for _, out := range []string{fmt.Sprint(s), fmt.Sprintf("%s %v", s, s), s.String()} {
		if out != "[redacted]" && out != "[redacted] [redacted]" {
			t.Errorf("printing a Secret gave %q, want [redacted]", out)
		}
	}
	if s.Reveal() != "hunter2-mpesa-pin" {
		t.Errorf("Reveal() = %q, want the real value", s.Reveal())
	}
	j, err := json.Marshal(map[string]Secret{"pin": s})
	if err != nil || string(j) != `{"pin":"[redacted]"}` {
		t.Errorf("json.Marshal = %s, %v; want {\"pin\":\"[redacted]\"}", j, err)
	}
}

func TestConfigLogsNoSecrets(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	log.Info("starting", "cfg", cfg)
	out := buf.String()
	checkNoLeak(t, "slog", out)
	for _, want := range []string{`"port":8080`, `postgres://gopher:xxxxx@db.internal:5432/matatu?sslmode=require`} {
		if !strings.Contains(out, want) {
			t.Errorf("the log should contain %s to help debugging:\n%s", want, out)
		}
	}

	buf.Reset()
	log.Info("jwt", "secret", cfg.JWTSecret)
	checkNoLeak(t, "slog of one Secret", buf.String())

	checkNoLeak(t, "fmt %v", fmt.Sprintf("%v", cfg))
	checkNoLeak(t, "fmt %+v", fmt.Sprintf("%+v", cfg))
	j, _ := json.Marshal(cfg)
	checkNoLeak(t, "json", string(j))
}

func TestRedactURL(t *testing.T) {
	tests := []struct{ raw, want string }{
		{"postgres://gopher:pw@localhost:5432/matatu", "postgres://gopher:xxxxx@localhost:5432/matatu"},
		{"postgres://localhost/matatu", "postgres://localhost/matatu"},
		{"postgres://gopher@localhost/matatu", "postgres://gopher@localhost/matatu"},
		{"postgres://u:p@host:bad port/db", "[invalid url]"},
	}
	for _, tt := range tests {
		if got := RedactURL(tt.raw); got != tt.want {
			t.Errorf("RedactURL(%q)\n got: %q\nwant: %q", tt.raw, got, tt.want)
		}
	}
}
