package runner

import (
	"testing"
)

func resetProxyFlags() { proxyURL, proxyAuth = "", "" }

func TestBuildProxyURL_UnsetIsNoOp(t *testing.T) {
	resetProxyFlags()
	u, err := buildProxyURL()
	if err != nil || u != nil {
		t.Fatalf("want (nil, nil) so ProxyFromEnvironment still applies, got (%v, %v)", u, err)
	}
}

func TestBuildProxyURL_Socks5WithAuth(t *testing.T) {
	resetProxyFlags()
	proxyURL, proxyAuth = "socks5://127.0.0.1:1080", "alice:s3cret"
	u, err := buildProxyURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme != "socks5" || u.Host != "127.0.0.1:1080" {
		t.Fatalf("unexpected URL: %s", u)
	}
	pass, ok := u.User.Password()
	if u.User.Username() != "alice" || !ok || pass != "s3cret" {
		t.Fatalf("credentials not applied: %v", u.User)
	}
}

// curl spells remote-resolution socks5h. http.Transport resolves at the proxy
// for socks5 anyway but rejects the scheme, so a URL copied from a working
// curl command would fail at request time with an opaque error.
func TestBuildProxyURL_Socks5hIsAccepted(t *testing.T) {
	resetProxyFlags()
	proxyURL = "socks5h://127.0.0.1:1080"
	u, err := buildProxyURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme != "socks5" {
		t.Fatalf("socks5h should normalise to socks5, got %q", u.Scheme)
	}
}

func TestBuildProxyURL_Rejects(t *testing.T) {
	for name, tc := range map[string]struct{ url, auth string }{
		"auth without proxy": {"", "alice:s3cret"},
		"unsupported scheme": {"ftp://127.0.0.1:1080", ""},
		"no host":            {"socks5://", ""},
		"auth without colon": {"socks5://127.0.0.1:1080", "alice"},
	} {
		t.Run(name, func(t *testing.T) {
			resetProxyFlags()
			proxyURL, proxyAuth = tc.url, tc.auth
			if _, err := buildProxyURL(); err == nil {
				t.Fatalf("expected an error for %s", name)
			}
		})
	}
}
