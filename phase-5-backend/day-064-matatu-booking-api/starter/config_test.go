package matatubookingapi

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var configVars = []string{"PORT", "JWT_SECRET", "JWT_ISSUER", "ACCESS_TOKEN_TTL", "CORS_ALLOWED_ORIGINS",
	"RATE_LIMIT_RPS", "RATE_LIMIT_BURST", "BCRYPT_COST", "ADMIN_PHONES", "CLIENT_IP_HEADER", "SHUTDOWN_TIMEOUT"}

func setConfigEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range configVars {
		t.Setenv(k, kv[k])
	}
}

func TestLoadConfig(t *testing.T) {
	setConfigEnv(t, map[string]string{
		"JWT_SECRET":           testSecret,
		"CORS_ALLOWED_ORIGINS": "https://matatu.example, http://localhost:5173",
		"ADMIN_PHONES":         "+254 712 345 678",
		"BCRYPT_COST":          "10",
		"CLIENT_IP_HEADER":     "Fly-Client-IP",
	})
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Port != 8080 || cfg.Issuer != "matatu-api" || cfg.AccessTTL != 15*time.Minute ||
		cfg.RatePerSecond != 5 || cfg.RateBurst != 20 || cfg.BcryptCost != 10 || cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("LoadConfig defaults = %+v", cfg)
	}
	if !reflect.DeepEqual(cfg.AllowedOrigins, []string{"https://matatu.example", "http://localhost:5173"}) {
		t.Errorf("AllowedOrigins = %q", cfg.AllowedOrigins)
	}
	if !reflect.DeepEqual(cfg.AdminPhones, []string{"0712345678"}) {
		t.Errorf("AdminPhones = %q, want [0712345678]", cfg.AdminPhones)
	}
	if cfg.ClientIPHeader != "Fly-Client-IP" {
		t.Errorf("ClientIPHeader = %q, want Fly-Client-IP", cfg.ClientIPHeader)
	}
	if string(cfg.JWTSecret) != testSecret || cfg.Now == nil {
		t.Errorf("JWTSecret or Now not set")
	}
}

func TestLoadConfigErrors(t *testing.T) {
	setConfigEnv(t, map[string]string{
		"PORT": "99999", "JWT_SECRET": "short", "ACCESS_TOKEN_TTL": "1d", "CORS_ALLOWED_ORIGINS": "*",
		"RATE_LIMIT_RPS": "0", "RATE_LIMIT_BURST": "none", "BCRYPT_COST": "40", "ADMIN_PHONES": "12345",
	})
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig with every value wrong: error = nil")
	}
	for _, k := range []string{"PORT", "JWT_SECRET", "ACCESS_TOKEN_TTL", "CORS_ALLOWED_ORIGINS", "RATE_LIMIT_RPS",
		"RATE_LIMIT_BURST", "BCRYPT_COST", "ADMIN_PHONES"} {
		if !strings.Contains(err.Error(), k) {
			t.Errorf("the error doesn't mention %s:\n%v", k, err)
		}
	}
	if strings.Contains(err.Error(), "short") {
		t.Errorf("the error contains the secret's value:\n%v", err)
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		in, want string
		ok       bool
	}{
		{"0712345678", "0712345678", true},
		{"0110345678", "0110345678", true},
		{"+254712345678", "0712345678", true},
		{"254712345678", "0712345678", true},
		{"0712 345 678", "0712345678", true},
		{"071234567", "", false},
		{"07123456789", "", false},
		{"0812345678", "", false},
		{"07123x5678", "", false},
		{"+255712345678", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalizePhone(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("NormalizePhone(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
