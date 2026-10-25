package secretsmanagement

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

const darajaKey = "sk_live_Wm9tYmEtbm90LXJlYWw"

func TestSecretIsRedacted(t *testing.T) {
	s := Secret(darajaKey)
	cfg := struct {
		User     string
		Password Secret
	}{"twende", s}
	for _, format := range []string{"%s", "%v", "%+v", "%#v", "%q"} {
		var out string
		if format == "%+v" || format == "%#v" {
			out = fmt.Sprintf(format, cfg)
		} else {
			out = fmt.Sprintf(format, s)
		}
		if strings.Contains(out, darajaKey) {
			t.Errorf("fmt.Sprintf(%q) printed the secret: %s", format, out)
		}
		if !strings.Contains(out, Redacted) {
			t.Errorf("fmt.Sprintf(%q) = %s, want it to show %s", format, out, Redacted)
		}
	}
	for _, h := range []string{"json", "text"} {
		var buf bytes.Buffer
		var log *slog.Logger
		if h == "json" {
			log = slog.New(slog.NewJSONHandler(&buf, nil))
		} else {
			log = slog.New(slog.NewTextHandler(&buf, nil))
		}
		log.Info("loaded config", "daraja_key", s, slog.Group("db", "password", s))
		if strings.Contains(buf.String(), darajaKey) || !strings.Contains(buf.String(), Redacted) {
			t.Errorf("%s log = %s, want %s and no secret", h, buf.String(), Redacted)
		}
	}
	if s.Reveal() != darajaKey {
		t.Errorf("Reveal() = %q, want the real value", s.Reveal())
	}
}

func env(vars map[string]string) Env {
	return Env{Prefix: "TWENDE_", Lookup: func(k string) (string, bool) { v, ok := vars[k]; return v, ok }}
}

func TestEnv(t *testing.T) {
	e := env(map[string]string{"TWENDE_DB_PASSWORD": "pg-pass", "TWENDE_EMPTY": "", "DB_PASSWORD": "wrong"})
	got, err := e.Get(context.Background(), "DB_PASSWORD")
	if err != nil || got.Reveal() != "pg-pass" {
		t.Errorf("Get(DB_PASSWORD) = %q, %v, want pg-pass from TWENDE_DB_PASSWORD", got.Reveal(), err)
	}
	for _, name := range []string{"MISSING", "EMPTY"} {
		_, err := e.Get(context.Background(), name)
		if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "TWENDE_"+name) {
			t.Errorf("Get(%s) = %v, want ErrNotFound naming TWENDE_%s", name, err, name)
		}
	}
	t.Setenv("TWENDE_TEST_ONLY", "from-os")
	if got, err := (Env{Prefix: "TWENDE_"}).Get(context.Background(), "TEST_ONLY"); err != nil || got.Reveal() != "from-os" {
		t.Errorf("with no Lookup, Get = %q, %v, want from-os via os.LookupEnv", got.Reveal(), err)
	}
}

func TestFiles(t *testing.T) {
	f := Files{FS: fstest.MapFS{
		"db_password":   {Data: []byte("pg-pass\n")},
		"windows":       {Data: []byte("abc\r\n")},
		"no_newline":    {Data: []byte("xyz")},
		"two_newlines":  {Data: []byte("keep\n\n")},
		"inner_spaces":  {Data: []byte("  padded  \n")},
		"sub/hidden":    {Data: []byte("nope")},
		"daraja_secret": {Data: []byte(darajaKey + "\n")},
	}}
	tests := []struct{ name, want string }{
		{"db_password", "pg-pass"},
		{"windows", "abc"},
		{"no_newline", "xyz"},
		{"two_newlines", "keep\n"},
		{"inner_spaces", "  padded  "},
		{"daraja_secret", darajaKey},
	}
	for _, tt := range tests {
		got, err := f.Get(context.Background(), tt.name)
		if err != nil || got.Reveal() != tt.want {
			t.Errorf("Get(%q) = %q, %v, want %q", tt.name, got.Reveal(), err, tt.want)
		}
	}
	if _, err := f.Get(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(missing) = %v, want ErrNotFound", err)
	}
	for _, bad := range []string{"", ".", "..", "../etc/passwd", "sub/hidden", `..\x`} {
		if _, err := f.Get(context.Background(), bad); !errors.Is(err, ErrBadName) {
			t.Errorf("Get(%q) = %v, want ErrBadName", bad, err)
		}
	}
}

// broken is a Provider that always fails with err.
type broken struct{ err error }

func (b broken) Get(context.Context, string) (Secret, error) { return "", b.err }

func TestChain(t *testing.T) {
	envP := env(map[string]string{"TWENDE_JWT_KEY": "from-env"})
	files := Files{FS: fstest.MapFS{"JWT_KEY": {Data: []byte("from-file")}, "DB_PASSWORD": {Data: []byte("pg-file")}}}
	c := Chain{envP, files}
	tests := []struct{ name, want string }{
		{"JWT_KEY", "from-env"},
		{"DB_PASSWORD", "pg-file"},
	}
	for _, tt := range tests {
		got, err := c.Get(context.Background(), tt.name)
		if err != nil || got.Reveal() != tt.want {
			t.Errorf("Chain.Get(%s) = %q, %v, want %q", tt.name, got.Reveal(), err, tt.want)
		}
	}
	if _, err := c.Get(context.Background(), "NOPE"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Chain.Get(NOPE) = %v, want ErrNotFound", err)
	}
	stop := Chain{broken{ErrForbidden}, files}
	if _, err := stop.Get(context.Background(), "DB_PASSWORD"); !errors.Is(err, ErrForbidden) {
		t.Errorf("Chain with a forbidden source = %v, want ErrForbidden: don't fall through to the next source", err)
	}
	if _, err := (Chain{}).Get(context.Background(), "X"); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty Chain = %v, want ErrNotFound", err)
	}
}

func TestVaultLeastPrivilege(t *testing.T) {
	v := NewFakeVault()
	if n := v.Put("twende/booking-api/DB_PASSWORD", "old"); n != 1 {
		t.Errorf("first Put = version %d, want 1", n)
	}
	if n := v.Put("twende/booking-api/DB_PASSWORD", "new"); n != 2 {
		t.Errorf("second Put = version %d, want 2", n)
	}
	v.Put("twende/payments/DARAJA_KEY", darajaKey)
	v.Put("twende/booking-api-admin/ROOT", "root")
	v.Grant("tok-booking", "twende/booking-api/*")
	v.Grant("tok-payments", "twende/payments/DARAJA_KEY")

	tests := []struct {
		token, path, want string
		err               error
	}{
		{"tok-booking", "twende/booking-api/DB_PASSWORD", "new", nil},
		{"tok-booking", "twende/booking-api/MISSING", "", ErrNotFound},
		{"tok-booking", "twende/payments/DARAJA_KEY", "", ErrForbidden},
		{"tok-booking", "twende/booking-api-admin/ROOT", "", ErrForbidden}, // a prefix of the name, not the folder
		{"tok-booking", "twende/booking-api", "", ErrForbidden},
		{"tok-payments", "twende/payments/DARAJA_KEY", darajaKey, nil},
		{"tok-payments", "twende/payments/OTHER", "", ErrForbidden},
		{"tok-nobody", "twende/payments/NOT_THERE", "", ErrForbidden}, // forbidden, not "not found": no probing
	}
	for _, tt := range tests {
		got, err := v.Read(tt.token, tt.path)
		if !errors.Is(err, tt.err) || got.Reveal() != tt.want {
			t.Errorf("Read(%s, %s) = %q, %v, want %q, %v", tt.token, tt.path, got.Reveal(), err, tt.want, tt.err)
		}
	}
	if a := v.Audit(); len(a) != len(tests) || a[2] != "tok-booking twende/payments/DARAJA_KEY" {
		t.Errorf("Audit = %q, want one entry per read attempt, denied ones included", a)
	}

	c := v.Client("tok-booking", "twende/booking-api")
	if got, err := c.Get(context.Background(), "DB_PASSWORD"); err != nil || got.Reveal() != "new" {
		t.Errorf("Client.Get = %q, %v, want new", got.Reveal(), err)
	}
}

func TestKeyringSignVerify(t *testing.T) {
	r := NewKeyring(Key{ID: "k1", Secret: "first-secret"})
	sig := r.Sign("booking:1042")
	// The HMAC-SHA256 of "booking:1042" with "first-secret", base64url without padding.
	if want := "k1.yG2qul-fmUXu4bjdm-dTPzkknrbXpBy4agnJ-gcbrIc"; sig != want {
		t.Errorf("Sign = %q, want %q", sig, want)
	}
	if err := r.Verify("booking:1042", sig); err != nil {
		t.Errorf("Verify(own signature) = %v, want nil", err)
	}
	bad := []string{
		"",
		"no-dot",
		"k1.",
		"k9." + strings.SplitN(sig, ".", 2)[1],
		sig + "x",
	}
	for _, s := range bad {
		if err := r.Verify("booking:1042", s); !errors.Is(err, ErrBadSignature) {
			t.Errorf("Verify(%q) = %v, want ErrBadSignature", s, err)
		}
	}
	if err := r.Verify("booking:9999", sig); !errors.Is(err, ErrBadSignature) {
		t.Errorf("Verify(other message) = %v, want ErrBadSignature", err)
	}
}

func TestKeyringRotation(t *testing.T) {
	r := NewKeyring(Key{ID: "k1", Secret: "first-secret"})
	old := r.Sign("ticket:7")
	r.Rotate(Key{ID: "k2", Secret: "second-secret"})
	fresh := r.Sign("ticket:7")
	if !strings.HasPrefix(fresh, "k2.") {
		t.Errorf("after Rotate, Sign = %q, want the new key k2", fresh)
	}
	if err := r.Verify("ticket:7", old); err != nil {
		t.Errorf("after one Rotate, an old-key signature = %v, want nil: keep accepting the previous key", err)
	}
	if err := r.Verify("ticket:7", fresh); err != nil {
		t.Errorf("new-key signature = %v, want nil", err)
	}
	r.Rotate(Key{ID: "k3", Secret: "third-secret"})
	if err := r.Verify("ticket:7", old); !errors.Is(err, ErrBadSignature) {
		t.Errorf("after two Rotates, a k1 signature = %v, want ErrBadSignature: k1 is retired", err)
	}
	if err := r.Verify("ticket:7", fresh); err != nil {
		t.Errorf("k2 signature after rotating to k3 = %v, want nil", err)
	}
}

func TestKeyringIsSafeForGoroutines(t *testing.T) {
	r := NewKeyring(Key{ID: "k0", Secret: "s0"})
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() { r.Rotate(Key{ID: fmt.Sprintf("k%d", i+1), Secret: Secret(fmt.Sprint("s", i+1))}) })
		wg.Go(func() { r.Verify("m", r.Sign("m")) })
	}
	wg.Wait()
	if err := r.Verify("m", r.Sign("m")); err != nil {
		t.Errorf("Verify after concurrent rotations = %v", err)
	}
}
