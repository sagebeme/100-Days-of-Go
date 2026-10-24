package passwords

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// cheap keeps the tests fast. Real code uses DefaultParams.
var cheap = Params{Memory: 64, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}

var blocklist = map[string]bool{
	"passwordpassword": true,
	"123456789012345":  true,
	"qwertyuiopasdfgh": true,
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     error
	}{
		{"15 characters", "matatu-ya-saa-6", nil},
		{"long passphrase, no digits or symbols", "nimechoka na jam ya thika road kila asubuhi", nil},
		{"14 characters", "matatu-ya-saa6", ErrTooShort},
		{"empty", "", ErrTooShort},
		{"64 characters", strings.Repeat("a", 64), nil},
		{"65 characters", strings.Repeat("a", 65), ErrTooLong},
		{"counted in characters, not bytes", strings.Repeat("é", 15), nil},
		{"14 two-byte characters is still too short", strings.Repeat("é", 14), ErrTooShort},
		{"blocklisted", "passwordpassword", ErrCommon},
		{"blocklist ignores case", "PasswordPassword", ErrCommon},
		{"spaces are allowed", "kuna    nafasi  hapa", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePassword(tt.password, blocklist)
			if !errors.Is(got, tt.want) {
				t.Errorf("ValidatePassword(%q)\n got: %v\nwant: %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestHashArgon2idFormat(t *testing.T) {
	h, err := HashArgon2id("mama-mboga-gikomba", DefaultParams)
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	prefix := "$argon2id$v=19$m=19456,t=2,p=1$"
	if !strings.HasPrefix(h, prefix) {
		t.Errorf("HashArgon2id\n got: %q\nwant a string starting %q", h, prefix)
	}
	if strings.Count(h, "$") != 5 {
		t.Errorf("HashArgon2id = %q, want 5 '$' separators", h)
	}
	if strings.Contains(h, "mama-mboga-gikomba") {
		t.Errorf("the hash contains the plain password: %q", h)
	}
}

func TestHashArgon2idSalts(t *testing.T) {
	a, err1 := HashArgon2id("same password twice", cheap)
	b, err2 := HashArgon2id("same password twice", cheap)
	if err1 != nil || err2 != nil {
		t.Fatalf("HashArgon2id errors: %v, %v", err1, err2)
	}
	if a == b || a == "" {
		t.Errorf("two hashes of the same password are equal (%q): each needs a fresh random salt", a)
	}
}

func TestVerifyArgon2id(t *testing.T) {
	h, err := HashArgon2id("kanjo wameshika kanjo", cheap)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		password string
		want     bool
	}{
		{"kanjo wameshika kanjo", true},
		{"kanjo wameshika kanjO", false},
		{"", false},
		{"kanjo wameshika kanjo ", false},
	}
	for _, tt := range tests {
		got, err := VerifyArgon2id(h, tt.password)
		if err != nil || got != tt.want {
			t.Errorf("VerifyArgon2id(hash, %q)\n got: %v, %v\nwant: %v, <nil>", tt.password, got, err, tt.want)
		}
	}
}

func TestVerifyArgon2idMalformed(t *testing.T) {
	bad := []string{
		"",
		"plain-text-password",
		"$argon2id$v=19$m=64,t=1,p=1$c2FsdA",
		"$argon2i$v=19$m=64,t=1,p=1$c2FsdHNhbHQ$a2V5a2V5",
		"$argon2id$v=19$m=abc,t=1,p=1$c2FsdHNhbHQ$a2V5a2V5",
		"$argon2id$v=19$m=64,t=1,p=1$!!!$a2V5a2V5",
	}
	for _, h := range bad {
		ok, err := VerifyArgon2id(h, "anything at all here")
		if ok || !errors.Is(err, ErrMalformedHash) {
			t.Errorf("VerifyArgon2id(%q)\n got: %v, %v\nwant: false, ErrMalformedHash", h, ok, err)
		}
	}
}

func TestVerifyAndRehash(t *testing.T) {
	const pw = "twende town kwa ile ya saa nane"
	current, _ := HashArgon2id(pw, cheap)
	weaker := cheap
	weaker.Time = 1
	weaker.Memory = 32
	old, _ := HashArgon2id(pw, weaker)
	legacy, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		hash       string
		password   string
		wantOK     bool
		wantRehash bool
	}{
		{"current argon2id, right password", current, pw, true, false},
		{"current argon2id, wrong password", current, "wrong password entirely", false, false},
		{"weaker argon2id, right password", old, pw, true, true},
		{"weaker argon2id, wrong password", old, "wrong password entirely", false, false},
		{"bcrypt from the old system", string(legacy), pw, true, true},
		{"bcrypt, wrong password", string(legacy), "wrong password entirely", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, rehash, err := Verify(tt.hash, tt.password, cheap)
			if err != nil || ok != tt.wantOK || rehash != tt.wantRehash {
				t.Errorf("Verify\n got: ok=%v rehash=%v err=%v\nwant: ok=%v rehash=%v err=<nil>", ok, rehash, err, tt.wantOK, tt.wantRehash)
			}
		})
	}

	if _, _, err := Verify("md5:5f4dcc3b5aa765d61d8327deb882cf99", pw, cheap); !errors.Is(err, ErrMalformedHash) {
		t.Errorf("Verify(unknown format) err = %v, want ErrMalformedHash", err)
	}
}

func TestHashBcrypt(t *testing.T) {
	h, err := HashBcrypt("sukuma wiki na ugali", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashBcrypt: %v", err)
	}
	if cost, err := bcrypt.Cost([]byte(h)); err != nil || cost != bcrypt.MinCost {
		t.Errorf("bcrypt.Cost(HashBcrypt(..., 4)) = %d, %v; want 4, <nil>", cost, err)
	}
	if bcrypt.CompareHashAndPassword([]byte(h), []byte("sukuma wiki na ugali")) != nil {
		t.Errorf("HashBcrypt output doesn't verify with bcrypt.CompareHashAndPassword")
	}
	if _, err := HashBcrypt(strings.Repeat("a", 73), bcrypt.MinCost); err == nil {
		t.Errorf("HashBcrypt(73 bytes) err = nil, want an error: bcrypt only reads 72 bytes")
	}
}

func newService(log *bytes.Buffer) (*Service, *MemStore) {
	store := NewMemStore()
	return &Service{
		Store:     store,
		Params:    cheap,
		Blocklist: blocklist,
		Log:       slog.New(slog.NewTextHandler(log, nil)),
	}, store
}

func TestRegisterAndLogin(t *testing.T) {
	var log bytes.Buffer
	s, store := newService(&log)
	const phone, pw = "0712345678", "ngong road ni jam tupu"

	if err := s.Register(phone, pw); err != nil {
		t.Fatalf("Register: %v", err)
	}
	u, ok := store.Get(phone)
	if !ok {
		t.Fatalf("Register didn't save the user")
	}
	if u.PasswordHash == pw || !strings.HasPrefix(u.PasswordHash, "$argon2id$") {
		t.Errorf("stored PasswordHash = %q, want an argon2id hash, never the password", u.PasswordHash)
	}
	if err := s.Register(phone, "another long password"); !errors.Is(err, ErrPhoneTaken) {
		t.Errorf("Register(same phone) = %v, want ErrPhoneTaken", err)
	}
	if err := s.Register("0722000000", "short"); !errors.Is(err, ErrTooShort) {
		t.Errorf("Register(short password) = %v, want ErrTooShort", err)
	}

	tests := []struct {
		phone, password string
		want            error
	}{
		{phone, pw, nil},
		{phone, "ngong road ni jam tupu!", ErrInvalidCredentials},
		{"0799999999", pw, ErrInvalidCredentials},
	}
	for _, tt := range tests {
		if got := s.Login(tt.phone, tt.password); !errors.Is(got, tt.want) {
			t.Errorf("Login(%q, %q)\n got: %v\nwant: %v", tt.phone, tt.password, got, tt.want)
		}
	}

	if strings.Contains(log.String(), pw) {
		t.Errorf("the log contains the plain password:\n%s", log.String())
	}
	if strings.Contains(log.String(), u.PasswordHash) {
		t.Errorf("the log contains the password hash:\n%s", log.String())
	}
	if !strings.Contains(log.String(), "login failed") {
		t.Errorf("failed logins should be logged as \"login failed\", got log:\n%s", log.String())
	}
}

func TestLoginRehashesOldHashes(t *testing.T) {
	var log bytes.Buffer
	s, store := newService(&log)
	const phone, pw = "0711111111", "mzee wa kanjo hana huruma"
	legacy, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store.Put(User{Phone: phone, PasswordHash: string(legacy)})

	if err := s.Login(phone, pw); err != nil {
		t.Fatalf("Login with a bcrypt hash: %v", err)
	}
	u, _ := store.Get(phone)
	if !strings.HasPrefix(u.PasswordHash, "$argon2id$") {
		t.Errorf("after login the stored hash is %q, want it replaced with argon2id", u.PasswordHash)
	}
	if err := s.Login(phone, pw); err != nil {
		t.Errorf("Login after the rehash: %v", err)
	}

	// A wrong password must never trigger a rehash.
	store.Put(User{Phone: phone, PasswordHash: string(legacy)})
	_ = s.Login(phone, "not the password at all")
	if u, _ := store.Get(phone); u.PasswordHash != string(legacy) {
		t.Errorf("a failed login changed the stored hash")
	}
}
