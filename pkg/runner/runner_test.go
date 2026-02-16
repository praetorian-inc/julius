package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTLSConfig_NilWhenNoFlagsSet(t *testing.T) {
	// Save original values
	origInsecure := insecureSkipVerify
	origCACert := caCertFile
	defer func() {
		insecureSkipVerify = origInsecure
		caCertFile = origCACert
	}()

	// Test: no flags set
	insecureSkipVerify = false
	caCertFile = ""

	tlsConfig, err := buildTLSConfig()
	require.NoError(t, err)
	assert.Nil(t, tlsConfig, "buildTLSConfig should return nil when no flags are set")
}

func TestBuildTLSConfig_WithInsecureFlag(t *testing.T) {
	// Save original values
	origInsecure := insecureSkipVerify
	origCACert := caCertFile
	defer func() {
		insecureSkipVerify = origInsecure
		caCertFile = origCACert
	}()

	insecureSkipVerify = true
	caCertFile = ""

	tlsConfig, err := buildTLSConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsConfig, "buildTLSConfig should return config when insecure flag set")
	assert.True(t, tlsConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
}

func TestBuildTLSConfig_InvalidCACertFile(t *testing.T) {
	// Save original values
	origInsecure := insecureSkipVerify
	origCACert := caCertFile
	defer func() {
		insecureSkipVerify = origInsecure
		caCertFile = origCACert
	}()

	insecureSkipVerify = false
	caCertFile = "/nonexistent/path/to/cert.pem"

	_, err := buildTLSConfig()
	assert.Error(t, err, "buildTLSConfig should return error for invalid CA cert file")
}

func TestBuildTLSConfig_ValidCACertFile(t *testing.T) {
	// Save original values
	origInsecure := insecureSkipVerify
	origCACert := caCertFile
	defer func() {
		insecureSkipVerify = origInsecure
		caCertFile = origCACert
	}()

	// Create a temporary CA cert file with a valid self-signed cert
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "ca.pem")
	// Valid self-signed certificate (generated with: openssl req -x509 -newkey rsa:2048 -nodes -days 365 -subj "/CN=test")
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIC/zCCAeegAwIBAgIUXpSGR6tdBCoct0eghI4fF7gvVCIwDQYJKoZIhvcNAQEL
BQAwDzENMAsGA1UEAwwEdGVzdDAeFw0yNjAyMTQxNTQ0NTZaFw0yNzAyMTQxNTQ0
NTZaMA8xDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQC6QsGr39yxVTsMxFZvuL4jWUkopPDVjXQkNIHnayJgNS85f9eBJ6yFX16q
bS4bMh6Tgy8Wdl74LHbGQ4qhGfoJbu37GCI8gqbc0ne/6qJJWLrfh9N98gM6907L
HxF5KdDQVJLMF8SvPzR5A8QztmfxOa+Ds1fEzWPQ0vkHRzPPp2gWmtY6m9aPDlnV
976t/Nya9UWTIDu/Tjn8e5MkUkm3uvQjTastW94owlJxqtL631i8dxAoXPiMxmx0
jRfiA27xwbXOXG2/wdnPlDECPHDkcetov1HP+3Kc1UOvyWGccN11Gr/siy5sQPjJ
zXr5Iwdk1fAWDcoKcnC1GSRZBKl/AgMBAAGjUzBRMB0GA1UdDgQWBBQ0VAGxojF9
bn9ZoQwcXkuwuUJ6wzAfBgNVHSMEGDAWgBQ0VAGxojF9bn9ZoQwcXkuwuUJ6wzAP
BgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBngJFsDr121fIls/DB
sXT9XTRnC9sBKPwq+vBp+AR+KOh1EZJ9Hoyn0l4noA99L498Xo/lkr+8tEq/8e5+
KEAh91IhOFv+HVkjby2b3DTCiXi0LckCn9TO1M03091TmrKvQJDqm4QBE/JoUXKJ
n7GuU94IcgjxunD9Aoo9oaGfq3tvb/MLdPvHyyBaTKwshBhHI+vENBnvY0jw1RNV
O9qmUCC+jbnfDuXv9eEDeUCTa0mt3UsigJZjsm43EASpsp7DnPerhDiVo7q/LBZU
W+qmbNe6MWuCmsF6q6mqA/hjSzdjrK5Q5STYdK49may9W3sFvgvLhnRtYczuq7Hh
o96z
-----END CERTIFICATE-----`)
	err := os.WriteFile(certFile, certPEM, 0600)
	require.NoError(t, err)

	insecureSkipVerify = false
	caCertFile = certFile

	tlsConfig, err := buildTLSConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsConfig, "buildTLSConfig should return config when CA cert provided")
	assert.NotNil(t, tlsConfig.RootCAs, "RootCAs should be set")
}

func TestLoadProbes_ReturnsEmbeddedWhenNoDirSet(t *testing.T) {
	// Save original value
	origProbesDir := probesDir
	defer func() {
		probesDir = origProbesDir
	}()

	probesDir = ""

	probes, err := loadProbes()
	require.NoError(t, err, "loadProbes should not error when probesDir is empty")
	assert.NotEmpty(t, probes, "loadProbes should return embedded probes")
}

func TestLoadProbes_LoadsFromDirWhenSet(t *testing.T) {
	// Save original value
	origProbesDir := probesDir
	defer func() {
		probesDir = origProbesDir
	}()

	// Create temp directory with a probe
	tempDir := t.TempDir()
	probeFile := filepath.Join(tempDir, "test.yaml")
	probeYAML := []byte(`name: test-probe
category: test
type: http
requests:
  - path: /test
    method: GET
    match:
      - type: status
        value: 200
`)
	err := os.WriteFile(probeFile, probeYAML, 0600)
	require.NoError(t, err)

	probesDir = tempDir

	probes, err := loadProbes()
	require.NoError(t, err, "loadProbes should not error when loading from directory")
	require.NotEmpty(t, probes, "loadProbes should return probes from directory")
	assert.Equal(t, "test-probe", probes[0].Name)
}
