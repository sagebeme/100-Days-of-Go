package linkshortener

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0"},
		{9, "9"},
		{10, "A"},
		{35, "Z"},
		{36, "a"},
		{61, "z"},
		{62, "10"},
		{3843, "zz"},
		{125000, "WW8"},
		{1<<64 - 1, "LygHa16AHYF"},
	}
	for _, tt := range tests {
		if got := Encode(tt.n); got != tt.want {
			t.Errorf("Encode(%d) = %q, want %q", tt.n, got, tt.want)
		}
		got, err := Decode(tt.want)
		if err != nil || got != tt.n {
			t.Errorf("Decode(%q) = %d, %v, want %d, nil", tt.want, got, err, tt.n)
		}
	}
}

func TestDecodeRejects(t *testing.T) {
	for _, s := range []string{"", "ab-c", "héllo", "a b", "LygHa16AHYG", "zzzzzzzzzzzz"} {
		if _, err := Decode(s); !errors.Is(err, ErrBadCode) {
			t.Errorf("Decode(%q) error = %v, want ErrBadCode", s, err)
		}
	}
}

func TestValidCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"aB3xY9q", true},
		{"0", true},
		{"", false},
		{"links", true},
		{"abc.def", false},
		{"abc%2F", false},
		{strings.Repeat("a", 17), false},
	}
	for _, tt := range tests {
		if got := ValidCode(tt.code); got != tt.want {
			t.Errorf("ValidCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestRandomCode(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		c := RandomCode()
		if len(c) != CodeLength || !ValidCode(c) {
			t.Fatalf("RandomCode() = %q, want %d base62 characters", c, CodeLength)
		}
		seen[c] = true
	}
	if len(seen) < 199 {
		t.Errorf("200 calls to RandomCode gave only %d different codes", len(seen))
	}
}

func TestPublicAddr(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"196.201.214.200", true},
		{"8.8.8.8", true},
		{"2001:4860:4860::8888", true},
		{"127.0.0.1", false},
		{"10.1.2.3", false},
		{"172.16.0.1", false},
		{"172.31.255.255", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false},
		{"100.64.0.1", false},
		{"0.0.0.0", false},
		{"255.255.255.255", false},
		{"224.0.0.1", false},
		{"::1", false},
		{"fe80::1", false},
		{"fd00::1", false},
		{"::ffff:127.0.0.1", false},
		{"::ffff:10.0.0.1", false},
	}
	for _, tt := range tests {
		if got := PublicAddr(netip.MustParseAddr(tt.addr)); got != tt.want {
			t.Errorf("PublicAddr(%s) = %v, want %v", tt.addr, got, tt.want)
		}
	}
}

// fakeDNS answers from a fixed table and fails for anything else.
func fakeDNS(table map[string][]string) LookupFunc {
	return func(_ context.Context, host string) ([]netip.Addr, error) {
		ips, ok := table[host]
		if !ok {
			return nil, errors.New("no such host")
		}
		var out []netip.Addr
		for _, ip := range ips {
			out = append(out, netip.MustParseAddr(ip))
		}
		return out, nil
	}
}

var dns = fakeDNS(map[string][]string{
	"duka.co.ke":        {"196.201.214.200"},
	"two-faced.example": {"196.201.214.201", "10.0.0.5"},
	"internal.example":  {"192.168.0.10"},
})

func TestValidateURLAccepts(t *testing.T) {
	for _, raw := range []string{
		"https://duka.co.ke/sale?item=sufuria&size=5L",
		"http://duka.co.ke",
		"HTTPS://DUKA.CO.KE/Path",
		"https://196.201.214.200/menu",
		"https://[2001:4860:4860::8888]/",
	} {
		if _, err := ValidateURL(context.Background(), raw, dns); err != nil {
			t.Errorf("ValidateURL(%q) = %v, want no error", raw, err)
		}
	}
}

func TestValidateURLRefuses(t *testing.T) {
	for _, raw := range []string{
		"",
		"duka.co.ke/no-scheme",
		"javascript:alert(document.cookie)",
		"JavaScript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"ftp://duka.co.ke/file",
		"file:///etc/passwd",
		"https://",
		"https://user:pass@duka.co.ke/",
		"http://localhost:8080/admin",
		"http://LOCALHOST./",
		"http://api.localhost/",
		"http://127.0.0.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]:6379/",
		"http://10.0.0.1/",
		"https://internal.example/",
		"https://two-faced.example/",
		"https://does-not-resolve.example/",
		"https://duka.co.ke/" + strings.Repeat("a", MaxURLLength),
	} {
		_, err := ValidateURL(context.Background(), raw, dns)
		var bad *URLError
		if !errors.As(err, &bad) {
			t.Errorf("ValidateURL(%q) error = %v, want a *URLError", raw, err)
		}
	}
}
