// Package configloader reads a service's settings from env variables, checks them
// all at startup and reports every problem at once. It also reads .env files for
// local development, and prints the config without leaking secrets.
package configloader

import (
	"io"
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
	// Your code here. Collect problems in a []error and return errors.Join(errs...).
	return Config{}, nil
}

// FromMap returns a LookupFunc that reads from m.
func FromMap(m map[string]string) LookupFunc {
	// Your code here.
	return nil
}

// Layered returns a LookupFunc that asks each lookup in turn and returns the first
// value found. Pass the real environment first, so it beats a .env file:
//
//	Layered(os.LookupEnv, FromMap(fileValues))
func Layered(lookups ...LookupFunc) LookupFunc {
	// Your code here.
	return nil
}

// ParseDotEnv reads a .env file: one KEY=value per line. It skips blank lines and lines
// starting with #, allows "export " before the key, trims spaces around the key and the
// value, and removes one pair of matching quotes ("..." or '...') around the value.
// Everything after the first = is the value, so values may contain =.
//
// A line with no =, or with a key that is empty or has characters other than letters,
// digits and _, is an error that names the line number: `line 3: no "=" in "PORT"`.
func ParseDotEnv(r io.Reader) (map[string]string, error) {
	// Your code here.
	return nil, nil
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
	// Your code here.
	return ""
}
