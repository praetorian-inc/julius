package scanner

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
)

// WithTLSConfig and WithProxy each customise the transport. If either cloned
// http.DefaultTransport unconditionally it would discard the other, making the
// result depend on argument order in NewScanner.
func TestTransportOptionsComposeInEitherOrder(t *testing.T) {
	proxy, err := url.Parse("socks5://127.0.0.1:1080")
	if err != nil {
		t.Fatal(err)
	}
	tlsCfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // test fixture

	for name, s := range map[string]*Scanner{
		"tls then proxy": NewScanner(WithTLSConfig(tlsCfg), WithProxy(proxy)),
		"proxy then tls": NewScanner(WithProxy(proxy), WithTLSConfig(tlsCfg)),
	} {
		t.Run(name, func(t *testing.T) {
			tr, ok := s.client.Transport.(*http.Transport)
			if !ok {
				t.Fatal("expected an *http.Transport")
			}
			if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
				t.Error("TLS config was lost")
			}
			if tr.Proxy == nil {
				t.Fatal("proxy was lost")
			}
			got, err := tr.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}})
			if err != nil || got == nil || got.String() != proxy.String() {
				t.Errorf("proxy = %v, %v; want %s", got, err, proxy)
			}
		})
	}
}

// A nil URL must leave DefaultTransport's ProxyFromEnvironment intact so
// HTTP_PROXY/HTTPS_PROXY keep working for callers that rely on them.
func TestWithProxy_NilIsNoOp(t *testing.T) {
	s := NewScanner(WithProxy(nil))
	if s.client.Transport != nil {
		t.Error("a nil proxy should not install a transport")
	}
}
