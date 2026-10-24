package owasptop10

import (
	"log/slog"
)

// Secret is a string that never prints: not with fmt, not with slog, not as JSON.
type Secret string

const redacted = "[redacted]"

// String hides the value from fmt's %s and %v.
func (Secret) String() string {
	// Your code here
	return ""
}

// LogValue hides the value from log/slog.
func (Secret) LogValue() slog.Value {
	// Your code here
	return slog.Value{}
}

// MarshalJSON hides the value from encoding/json.
func (Secret) MarshalJSON() ([]byte, error) {
	// Your code here
	return nil, nil
}

// Reveal returns the real value, for the one place that needs it.
func (s Secret) Reveal() string { return string(s) }

// RedactURL returns raw with any password replaced by "xxxxx", or "[invalid url]" if
// it can't be parsed. A DATABASE_URL usually has the password inside it.
func RedactURL(raw string) string {
	// Your code here
	return ""
}

// Config is the API's settings.
type Config struct {
	Port        int
	DatabaseURL Secret
	JWTSecret   Secret
	MpesaKey    Secret
}

// LogValue lets slog.Info("config", "cfg", cfg) log what helps debugging, the port and
// the database URL with its password redacted, and nothing secret.
func (c Config) LogValue() slog.Value {
	// Your code here
	return slog.Value{}
}
