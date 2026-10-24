// Package configloader reads a service's settings from env variables, checks them
// all at startup and reports every problem at once. It also reads .env files for
// local development, and prints the config without leaking secrets.
package configloader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LookupFunc finds an env variable. It has the same shape as os.LookupEnv, so the
// real program passes os.LookupEnv and tests pass a function backed by a map.
type LookupFunc func(key string) (string, bool)

// Config holds every setting the chama reminder service needs.
type Config struct {
	Env            string        // APP_ENV: development, staging or production
	Port           int           // PORT
	DatabaseURL    string        // DATABASE_URL
	DarajaKey      string        // DARAJA_CONSUMER_KEY
	DarajaSecret   string        // DARAJA_CONSUMER_SECRET
	RequestTimeout time.Duration // REQUEST_TIMEOUT
	Debug          bool          // DEBUG
}

// Load builds a Config from lookup. A variable that is set to "" counts as not set.
//
//	APP_ENV                 optional, default "development"; one of development, staging, production
//	PORT                    optional, default 8080; a whole number from 1 to 65535
//	DATABASE_URL            required
//	DARAJA_CONSUMER_KEY     required
//	DARAJA_CONSUMER_SECRET  required
//	REQUEST_TIMEOUT         optional, default 5s; a Go duration above 0
//	DEBUG                   optional, default false; parsed with strconv.ParseBool
//
// DEBUG can't be true when APP_ENV is production. Load checks every variable, in the
// order above, and joins all the problems with errors.Join, one per line:
//
//	APP_ENV must be development, staging or production, got "prod"
//	PORT must be a whole number, got "80a"
//	PORT must be from 1 to 65535, got 70000
//	DATABASE_URL is required
//	REQUEST_TIMEOUT must be a duration like 5s or 1m, got "5"
//	REQUEST_TIMEOUT must be more than 0, got 0s
//	DEBUG must be true or false, got "yes"
//	DEBUG can't be true when APP_ENV is production
func Load(lookup LookupFunc) (Config, error) {
	get := func(key string) (string, bool) {
		v, ok := lookup(key)
		return v, ok && v != ""
	}
	var cfg Config
	var errs []error

	cfg.Env = "development"
	if v, ok := get("APP_ENV"); ok {
		switch v {
		case "development", "staging", "production":
			cfg.Env = v
		default:
			errs = append(errs, fmt.Errorf("APP_ENV must be development, staging or production, got %q", v))
		}
	}

	cfg.Port = 8080
	if v, ok := get("PORT"); ok {
		n, err := strconv.Atoi(v)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("PORT must be a whole number, got %q", v))
		case n < 1 || n > 65535:
			errs = append(errs, fmt.Errorf("PORT must be from 1 to 65535, got %d", n))
		default:
			cfg.Port = n
		}
	}

	required := []struct {
		key string
		dst *string
	}{
		{"DATABASE_URL", &cfg.DatabaseURL},
		{"DARAJA_CONSUMER_KEY", &cfg.DarajaKey},
		{"DARAJA_CONSUMER_SECRET", &cfg.DarajaSecret},
	}
	for _, r := range required {
		v, ok := get(r.key)
		if !ok {
			errs = append(errs, fmt.Errorf("%s is required", r.key))
			continue
		}
		*r.dst = v
	}

	cfg.RequestTimeout = 5 * time.Second
	if v, ok := get("REQUEST_TIMEOUT"); ok {
		d, err := time.ParseDuration(v)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("REQUEST_TIMEOUT must be a duration like 5s or 1m, got %q", v))
		case d <= 0:
			errs = append(errs, fmt.Errorf("REQUEST_TIMEOUT must be more than 0, got %s", d))
		default:
			cfg.RequestTimeout = d
		}
	}

	if v, ok := get("DEBUG"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("DEBUG must be true or false, got %q", v))
		}
		cfg.Debug = b
	}
	if cfg.Debug && cfg.Env == "production" {
		errs = append(errs, errors.New("DEBUG can't be true when APP_ENV is production"))
	}

	return cfg, errors.Join(errs...)
}

// FromMap returns a LookupFunc that reads from m.
func FromMap(m map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

// Layered returns a LookupFunc that asks each lookup in turn and returns the first
// value found. Pass the real environment first, so it beats a .env file:
//
//	Layered(os.LookupEnv, FromMap(fileValues))
func Layered(lookups ...LookupFunc) LookupFunc {
	return func(key string) (string, bool) {
		for _, l := range lookups {
			if v, ok := l(key); ok {
				return v, true
			}
		}
		return "", false
	}
}

// ParseDotEnv reads a .env file: one KEY=value per line. It skips blank lines and lines
// starting with #, allows "export " before the key, trims spaces around the key and the
// value, and removes one pair of matching quotes ("..." or '...') around the value.
// Everything after the first = is the value, so values may contain =.
//
// A line with no =, or with a key that is empty or has characters other than letters,
// digits and _, is an error that names the line number: `line 3: no "=" in "PORT"`.
func ParseDotEnv(r io.Reader) (map[string]string, error) {
	values := map[string]string{}
	sc := bufio.NewScanner(r)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: no \"=\" in %q", n, line)
		}
		key = strings.TrimSpace(key)
		if !validKey(key) {
			return nil, fmt.Errorf("line %d: bad key %q", n, key)
		}
		values[key] = unquote(strings.TrimSpace(value))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

// String prints the config one setting per line, in the order Load reads them, with
// secrets hidden. The Daraja key and secret show as "****". The database URL keeps
// everything except its password, which becomes "xxxxx" (url.URL.Redacted does
// that); a URL that doesn't parse shows as "****".
//
//	APP_ENV=production
//	PORT=8080
//	DATABASE_URL=postgres://chama:xxxxx@db.internal:5432/chama
//	DARAJA_CONSUMER_KEY=****
//	DARAJA_CONSUMER_SECRET=****
//	REQUEST_TIMEOUT=5s
//	DEBUG=false
func (c Config) String() string {
	dbURL := "****"
	if u, err := url.Parse(c.DatabaseURL); err == nil {
		dbURL = u.Redacted()
	}
	return fmt.Sprintf("APP_ENV=%s\nPORT=%d\nDATABASE_URL=%s\nDARAJA_CONSUMER_KEY=****\nDARAJA_CONSUMER_SECRET=****\nREQUEST_TIMEOUT=%s\nDEBUG=%t",
		c.Env, c.Port, dbURL, c.RequestTimeout, c.Debug)
}
