package configloader

import (
	"maps"
	"strings"
	"testing"
	"time"
)

// env is a fake environment for tests: a LookupFunc backed by a map. The tests never
// read or change the real environment, so they can't depend on your machine.
func env(m map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

// valid has every required variable and nothing else.
func valid() map[string]string {
	return map[string]string{
		"DATABASE_URL":           "postgres://chama:s3cret-pw@localhost:5432/chama",
		"DARAJA_CONSUMER_KEY":    "test-key-AbC123",
		"DARAJA_CONSUMER_SECRET": "test-secret-XyZ789",
	}
}

func with(changes map[string]string) map[string]string {
	m := valid()
	maps.Copy(m, changes)
	return m
}

func TestLoadDefaults(t *testing.T) {
	got, err := Load(env(valid()))
	if err != nil {
		t.Fatalf("Load(only the required variables) returned an error:\n%v", err)
	}
	want := Config{
		Env:            "development",
		Port:           8080,
		DatabaseURL:    "postgres://chama:s3cret-pw@localhost:5432/chama",
		DarajaKey:      "test-key-AbC123",
		DarajaSecret:   "test-secret-XyZ789",
		RequestTimeout: 5 * time.Second,
		Debug:          false,
	}
	if got != want {
		t.Errorf("Load(only the required variables)\n got: %#v\nwant: %#v", got, want)
	}
}

func TestLoadEverything(t *testing.T) {
	got, err := Load(env(with(map[string]string{
		"APP_ENV":         "staging",
		"PORT":            "9000",
		"REQUEST_TIMEOUT": "1m30s",
		"DEBUG":           "true",
	})))
	if err != nil {
		t.Fatalf("Load returned an error:\n%v", err)
	}
	if got.Env != "staging" || got.Port != 9000 || got.RequestTimeout != 90*time.Second || !got.Debug {
		t.Errorf("Load\n got: Env %q, Port %d, RequestTimeout %v, Debug %v\nwant: Env \"staging\", Port 9000, RequestTimeout 1m30s, Debug true",
			got.Env, got.Port, got.RequestTimeout, got.Debug)
	}
}

func TestLoadEmptyMeansUnset(t *testing.T) {
	got, err := Load(env(with(map[string]string{"PORT": "", "APP_ENV": "", "DEBUG": "", "REQUEST_TIMEOUT": ""})))
	if err != nil {
		t.Fatalf("Load with empty optional variables returned an error:\n%v", err)
	}
	if got.Port != 8080 || got.Env != "development" || got.RequestTimeout != 5*time.Second {
		t.Errorf("empty values should give the defaults; got Port %d, Env %q, RequestTimeout %v", got.Port, got.Env, got.RequestTimeout)
	}

	_, err = Load(env(with(map[string]string{"DATABASE_URL": ""})))
	if err == nil || err.Error() != "DATABASE_URL is required" {
		t.Errorf("Load with DATABASE_URL=\"\"\n got: %v\nwant: DATABASE_URL is required", err)
	}
}

func TestLoadOneProblem(t *testing.T) {
	tests := []struct {
		name    string
		changes map[string]string
		want    string
	}{
		{"bad APP_ENV", map[string]string{"APP_ENV": "prod"}, `APP_ENV must be development, staging or production, got "prod"`},
		{"PORT not a number", map[string]string{"PORT": "80a"}, `PORT must be a whole number, got "80a"`},
		{"PORT with a space", map[string]string{"PORT": "8080 "}, `PORT must be a whole number, got "8080 "`},
		{"PORT too big", map[string]string{"PORT": "70000"}, `PORT must be from 1 to 65535, got 70000`},
		{"PORT zero", map[string]string{"PORT": "0"}, `PORT must be from 1 to 65535, got 0`},
		{"timeout with no unit", map[string]string{"REQUEST_TIMEOUT": "5"}, `REQUEST_TIMEOUT must be a duration like 5s or 1m, got "5"`},
		{"timeout zero", map[string]string{"REQUEST_TIMEOUT": "0s"}, `REQUEST_TIMEOUT must be more than 0, got 0s`},
		{"negative timeout", map[string]string{"REQUEST_TIMEOUT": "-2s"}, `REQUEST_TIMEOUT must be more than 0, got -2s`},
		{"DEBUG yes", map[string]string{"DEBUG": "yes"}, `DEBUG must be true or false, got "yes"`},
		{"debug in production", map[string]string{"APP_ENV": "production", "DEBUG": "1"}, `DEBUG can't be true when APP_ENV is production`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(env(with(tt.changes)))
			if err == nil {
				t.Fatalf("Load(%v) = nil error, want: %s", tt.changes, tt.want)
			}
			if err.Error() != tt.want {
				t.Errorf("Load(%v)\n got: %s\nwant: %s", tt.changes, err, tt.want)
			}
		})
	}
}

func TestLoadReportsEverything(t *testing.T) {
	_, err := Load(env(map[string]string{
		"APP_ENV":         "production",
		"PORT":            "abc",
		"REQUEST_TIMEOUT": "ten seconds",
		"DEBUG":           "true",
	}))
	want := strings.Join([]string{
		`PORT must be a whole number, got "abc"`,
		`DATABASE_URL is required`,
		`DARAJA_CONSUMER_KEY is required`,
		`DARAJA_CONSUMER_SECRET is required`,
		`REQUEST_TIMEOUT must be a duration like 5s or 1m, got "ten seconds"`,
		`DEBUG can't be true when APP_ENV is production`,
	}, "\n")
	if err == nil || err.Error() != want {
		t.Errorf("Load with six problems should report all six, in order, one per line\n got:\n%v\n\nwant:\n%s", err, want)
	}
}

func TestFromMap(t *testing.T) {
	lookup := FromMap(map[string]string{"PORT": "9000", "EMPTY": ""})
	if lookup == nil {
		t.Fatal("FromMap returned nil, want a function")
	}
	tests := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"PORT", "9000", true},
		{"EMPTY", "", true},
		{"MISSING", "", false},
	}
	for _, tt := range tests {
		got, ok := lookup(tt.key)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("FromMap(...)(%q) = %q, %v; want %q, %v", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestLayered(t *testing.T) {
	real := env(map[string]string{"PORT": "9000", "DEBUG": ""})
	file := env(map[string]string{"PORT": "8081", "APP_ENV": "staging", "DEBUG": "true"})
	lookup := Layered(real, file)
	if lookup == nil {
		t.Fatal("Layered returned nil, want a function")
	}
	tests := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"PORT", "9000", true},       // the real environment wins
		{"APP_ENV", "staging", true}, // only in the file
		{"DEBUG", "", true},          // set to "" in the real environment still wins
		{"MISSING", "", false},
	}
	for _, tt := range tests {
		got, ok := lookup(tt.key)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("Layered(real, file)(%q) = %q, %v; want %q, %v", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
	if Layered() == nil {
		t.Fatal("Layered() returned nil, want a function")
	}
	if _, ok := Layered()("PORT"); ok {
		t.Errorf("Layered() with no lookups found PORT, want not found")
	}
}

func TestParseDotEnv(t *testing.T) {
	file := `# Local settings for the chama service
APP_ENV=development

PORT = 8081
export DEBUG=true
DATABASE_URL=postgres://chama:pw@localhost:5432/chama?sslmode=disable
DARAJA_CONSUMER_KEY="key with spaces"
DARAJA_CONSUMER_SECRET='s3cr=t'
   # an indented comment
EMPTY=
HALF_QUOTED="oops
`
	got, err := ParseDotEnv(strings.NewReader(file))
	if err != nil {
		t.Fatalf("ParseDotEnv returned an error: %v", err)
	}
	want := map[string]string{
		"APP_ENV":                "development",
		"PORT":                   "8081",
		"DEBUG":                  "true",
		"DATABASE_URL":           "postgres://chama:pw@localhost:5432/chama?sslmode=disable",
		"DARAJA_CONSUMER_KEY":    "key with spaces",
		"DARAJA_CONSUMER_SECRET": "s3cr=t",
		"EMPTY":                  "",
		"HALF_QUOTED":            `"oops`,
	}
	if !maps.Equal(got, want) {
		t.Errorf("ParseDotEnv\n got: %q\nwant: %q", got, want)
	}

	empty, err := ParseDotEnv(strings.NewReader(""))
	if err != nil || empty == nil || len(empty) != 0 {
		t.Errorf("ParseDotEnv(\"\") = %#v, %v; want an empty map and nil", empty, err)
	}
}

func TestParseDotEnvErrors(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string // the error must contain this
	}{
		{"no equals sign", "APP_ENV=development\nPORT\n", "line 2"},
		{"empty key", "# settings\n\n=8080\n", "line 3"},
		{"space in the key", "MY PORT=8080\n", "line 1"},
		{"dash in the key", "APP_ENV=dev\nDB-URL=x\n", "line 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDotEnv(strings.NewReader(tt.file))
			if err == nil {
				t.Fatalf("ParseDotEnv(%q) = %v, nil; want an error mentioning %q", tt.file, got, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseDotEnv(%q) error %q doesn't mention %q", tt.file, err, tt.want)
			}
		})
	}
}

func TestStringHidesSecrets(t *testing.T) {
	cfg := Config{
		Env:            "production",
		Port:           8080,
		DatabaseURL:    "postgres://chama:s3cret-pw@db.internal:5432/chama",
		DarajaKey:      "live-key-AbC123",
		DarajaSecret:   "live-secret-XyZ789",
		RequestTimeout: 5 * time.Second,
	}
	got := cfg.String()
	want := `APP_ENV=production
PORT=8080
DATABASE_URL=postgres://chama:xxxxx@db.internal:5432/chama
DARAJA_CONSUMER_KEY=****
DARAJA_CONSUMER_SECRET=****
REQUEST_TIMEOUT=5s
DEBUG=false`
	if got != want {
		t.Errorf("Config.String()\n got:\n%s\n\nwant:\n%s", got, want)
	}
	for _, secret := range []string{"s3cret-pw", "live-key-AbC123", "live-secret-XyZ789"} {
		if strings.Contains(got, secret) {
			t.Errorf("Config.String() leaks the secret %q", secret)
		}
	}

	cfg.DatabaseURL = "postgres://chama:pw@bad host:5432/chama"
	if got := cfg.String(); !strings.Contains(got, "DATABASE_URL=****\n") {
		t.Errorf("a DATABASE_URL that doesn't parse should show as ****; got:\n%s", got)
	}
}
