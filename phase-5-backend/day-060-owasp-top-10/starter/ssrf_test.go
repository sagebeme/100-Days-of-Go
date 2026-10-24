package owasptop10

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.8.8.8", true},
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // cloud metadata
		{"0.0.0.0", true},
		{"0.1.2.3", true},
		{"100.64.0.1", true},
		{"100.128.0.1", false},
		{"224.0.0.1", true},
		{"::1", true},
		{"::", true},
		{"fd00::1", true},
		{"fe80::1", true},
		{"::ffff:127.0.0.1", true},
		{"::ffff:10.1.1.1", true},
		{"196.201.214.200", false},
		{"8.8.8.8", false},
		{"2001:4860:4860::8888", false},
	}
	for _, tt := range tests {
		if got := IsBlockedIP(netip.MustParseAddr(tt.ip)); got != tt.want {
			t.Errorf("IsBlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
	if !IsBlockedIP(netip.Addr{}) {
		t.Errorf("IsBlockedIP(zero Addr) = false, want true")
	}
}

var allowed = map[string]bool{"api.safaricom.co.ke": true, "sandbox.safaricom.co.ke": true}

func TestCheckURL(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{"https://api.safaricom.co.ke/mpesa/stkpush/v1/processrequest", true},
		{"https://sandbox.safaricom.co.ke:443/oauth", true},
		{"http://api.safaricom.co.ke/", false},
		{"https://evil.example/", false},
		{"https://api.safaricom.co.ke.evil.example/", false},
		{"https://user:pass@api.safaricom.co.ke/", false},
		{"https://169.254.169.254/latest/meta-data/", false},
		{"file:///etc/passwd", false},
		{"gopher://api.safaricom.co.ke/", false},
		{"https://%zz", false},
		{"", false},
	}
	for _, tt := range tests {
		u, err := CheckURL(tt.raw, allowed)
		if tt.want && (err != nil || u == nil) {
			t.Errorf("CheckURL(%q) = %v, %v; want it allowed", tt.raw, u, err)
		}
		if !tt.want && (!errors.Is(err, ErrBlocked) || u != nil) {
			t.Errorf("CheckURL(%q) = %v, %v; want nil, ErrBlocked", tt.raw, u, err)
		}
	}
}

func TestSafeClientBlocksAtDialTime(t *testing.T) {
	hit := false
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true }))
	defer srv.Close()

	// Even with 127.0.0.1 on the allowlist, the dialer must refuse it.
	c := NewSafeClient(map[string]bool{"127.0.0.1": true}, 2*time.Second)
	resp, err := c.Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("GET %s succeeded; the safe client must refuse loopback addresses", srv.URL)
	}
	if !errors.Is(err, ErrBlocked) {
		t.Errorf("GET loopback error = %v, want one wrapping ErrBlocked", err)
	}
	if hit {
		t.Errorf("the request reached the server")
	}
	if c.Timeout != 2*time.Second {
		t.Errorf("client Timeout = %v, want 2s", c.Timeout)
	}
}

func TestSafeClientRedirects(t *testing.T) {
	c := NewSafeClient(allowed, time.Second)
	if c.CheckRedirect == nil {
		t.Fatal("CheckRedirect is nil: the client follows redirects anywhere")
	}
	req := func(u string) *http.Request { r, _ := http.NewRequest("GET", u, nil); return r }
	first := []*http.Request{req("https://api.safaricom.co.ke/a")}
	if err := c.CheckRedirect(req("https://sandbox.safaricom.co.ke/b"), first); err != nil {
		t.Errorf("redirect to an allowed host: %v, want nil", err)
	}
	if err := c.CheckRedirect(req("http://169.254.169.254/latest/meta-data/"), first); err == nil {
		t.Errorf("redirect to the metadata address was allowed")
	}
	three := []*http.Request{first[0], first[0], first[0]}
	if err := c.CheckRedirect(req("https://api.safaricom.co.ke/c"), three); err == nil {
		t.Errorf("a fourth redirect was allowed, want at most 3")
	}
}
