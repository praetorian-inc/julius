package sdk

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/praetorian-inc/capability-sdk/pkg/capability"
	"github.com/praetorian-inc/capability-sdk/pkg/capmodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mcpDiscoverResponse is a valid MCP server/discover JSON-RPC response
// per the 2026-07-28 spec. The scanner's match rules require the body
// to contain "supportedVersions" and "io.modelcontextprotocol/serverInfo".
const mcpDiscoverResponse = `{"jsonrpc":"2.0","id":1,"result":{"supportedVersions":["2026-07-28"],"_meta":{"io.modelcontextprotocol/serverInfo":{"name":"test-mcp","version":"1.0"}},"capabilities":{"tools":{"listChanged":true}}}}`

// startMCPServer creates an httptest TLS server that serves valid MCP
// discover responses to POST requests and 404 for everything else.
// The returned cleanup function restores http.DefaultTransport.
func startMCPServer(t *testing.T, response string, statusCode int) (*httptest.Server, string, int, func()) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprint(w, response)
	}))

	host, port := hostPort(t, srv)

	// Override DefaultTransport so the scanner's lazily-created http.Client
	// trusts the test server's self-signed TLS certificate.
	origTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport

	cleanup := func() {
		http.DefaultTransport = origTransport
		srv.Close()
	}

	return srv, host, port, cleanup
}

// hostPort extracts the host and numeric port from an httptest.Server URL.
func hostPort(t *testing.T, srv *httptest.Server) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(srv.Listener.Addr().String())
	require.NoError(t, err, "parsing test server address")
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	require.NoError(t, err, "parsing port number")
	return host, port
}

// collectEmitted returns an EmitterFunc that appends all emitted models to a
// slice. This tests the actual Invoke output -- not a mock.
func collectEmitted(collected *[]any) capability.EmitterFunc {
	return capability.EmitterFunc(func(models ...any) error {
		*collected = append(*collected, models...)
		return nil
	})
}

// TestInvoke_PositiveDetection verifies that Invoke emits a WebApplication and
// a Webpage when scanning an MCP endpoint that returns a valid discover response.
func TestInvoke_PositiveDetection(t *testing.T) {
	_, host, port, cleanup := startMCPServer(t, mcpDiscoverResponse, http.StatusOK)
	defer cleanup()

	cap := NewCapability()
	input := capmodel.Port{
		Service: "https",
		Port:    port,
		Parent:  capmodel.Asset{DNS: host},
	}

	var emitted []any
	err := cap.Invoke(capability.ExecutionContext{}, input, collectEmitted(&emitted))
	require.NoError(t, err, "Invoke should not return an error for a valid MCP server")

	require.Len(t, emitted, 2, "Invoke should emit exactly 2 models (WebApplication + Webpage)")

	// First model: WebApplication
	webApp, ok := emitted[0].(WebApplication)
	require.True(t, ok, "first emitted model should be WebApplication, got %T", emitted[0])
	expectedURL := fmt.Sprintf("https://%s:%d", host, port)
	assert.Equal(t, expectedURL, webApp.PrimaryURL, "WebApplication.PrimaryURL should be the target URL")
	assert.Equal(t, "mcp-server", webApp.Name, "WebApplication.Name should be the probe's service name")
	assert.Equal(t, "A", webApp.Status, "WebApplication.Status should be 'A' (active)")

	// Second model: Webpage
	webpage, ok := emitted[1].(Webpage)
	require.True(t, ok, "second emitted model should be Webpage, got %T", emitted[1])
	assert.Equal(t, expectedURL, webpage.URL, "Webpage.URL should be the target URL")
	assert.Equal(t, "mcp", webpage.Type, "Webpage.Type should be 'mcp' for MCP category probes")
	assert.Equal(t, "mcp-server", webpage.Service, "Webpage.Service should match the detected service")
	assert.Equal(t, 90, webpage.Specificity, "Webpage.Specificity should be 90 (from the mcp-server probe)")
}

// TestInvoke_GeneratorConfigs_AugustusCompatibility verifies that the
// GeneratorConfigs emitted by Invoke carry the fields Augustus needs to consume
// the detected endpoint. If GeneratorConfigs are empty or malformed,
// Augustus.Match() rejects the webpage.
func TestInvoke_GeneratorConfigs_AugustusCompatibility(t *testing.T) {
	_, host, port, cleanup := startMCPServer(t, mcpDiscoverResponse, http.StatusOK)
	defer cleanup()

	cap := NewCapability()
	input := capmodel.Port{
		Service: "https",
		Port:    port,
		Parent:  capmodel.Asset{DNS: host},
	}

	var emitted []any
	err := cap.Invoke(capability.ExecutionContext{}, input, collectEmitted(&emitted))
	require.NoError(t, err)
	require.Len(t, emitted, 2)

	webpage, ok := emitted[1].(Webpage)
	require.True(t, ok, "second emitted model should be Webpage")

	// Augustus requires non-empty GeneratorConfigs to accept a webpage.
	require.NotEmpty(t, webpage.GeneratorConfigs,
		"GeneratorConfigs must not be empty -- Augustus rejects webpages without them")

	cfg := webpage.GeneratorConfigs[0]

	// Source: toGeneratorConfigs stamps "julius" on every config.
	assert.Equal(t, "julius", cfg.Source,
		"GeneratorConfig.Source must be 'julius' so Augustus knows the provenance")

	// Type: the MCP probe's augustus.generator is "mcp".
	assert.Equal(t, "mcp", cfg.Type,
		"GeneratorConfig.Type must be 'mcp' (the Augustus generator type)")

	// Endpoint: must contain the target URL ($TARGET is substituted).
	expectedURL := fmt.Sprintf("https://%s:%d", host, port)
	assert.NotEmpty(t, cfg.Endpoint, "GeneratorConfig.Endpoint must not be empty")
	assert.Contains(t, cfg.Endpoint, expectedURL,
		"GeneratorConfig.Endpoint should contain the target URL")

	// Extra: MCP-specific transport and mode fields.
	require.NotNil(t, cfg.Extra, "GeneratorConfig.Extra must not be nil for MCP configs")
	assert.Equal(t, "streamable-http", cfg.Extra["transport"],
		"Extra['transport'] must be 'streamable-http' for MCP Streamable HTTP transport")
	assert.Equal(t, "list_tools", cfg.Extra["mode"],
		"Extra['mode'] must be 'list_tools' for Augustus MCP generator")
}

// TestInvoke_GeneratorConfigs_JSONRoundTrip verifies that the SDK
// GeneratorConfig survives a JSON marshal/unmarshal round-trip and that the
// JSON keys match what Guard's model.GeneratorConfig expects. A tag mismatch
// (e.g. "source" vs "Source") would cause silent data loss in the registry
// converter.
func TestInvoke_GeneratorConfigs_JSONRoundTrip(t *testing.T) {
	cfg := GeneratorConfig{
		Source:   "julius",
		Type:     "mcp",
		Endpoint: "https://example.com:8080",
		Extra:    map[string]string{"transport": "streamable-http", "mode": "list_tools"},
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err, "GeneratorConfig should marshal to JSON")

	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err, "JSON should unmarshal to map")

	// Verify expected JSON keys are present (these match Guard's model).
	assert.Equal(t, "julius", raw["source"], "JSON key 'source' should carry the source value")
	assert.Equal(t, "mcp", raw["type"], "JSON key 'type' should carry the generator type")
	assert.Equal(t, "https://example.com:8080", raw["endpoint"], "JSON key 'endpoint' should carry the endpoint URL")

	extraRaw, ok := raw["extra"].(map[string]any)
	require.True(t, ok, "JSON key 'extra' should be an object")
	assert.Equal(t, "streamable-http", extraRaw["transport"])
	assert.Equal(t, "list_tools", extraRaw["mode"])

	// Verify round-trip fidelity: unmarshal back to GeneratorConfig.
	var roundTripped GeneratorConfig
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err, "JSON should unmarshal back to GeneratorConfig")
	assert.Equal(t, cfg.Source, roundTripped.Source)
	assert.Equal(t, cfg.Type, roundTripped.Type)
	assert.Equal(t, cfg.Endpoint, roundTripped.Endpoint)
	assert.Equal(t, cfg.Extra["transport"], roundTripped.Extra["transport"])
	assert.Equal(t, cfg.Extra["mode"], roundTripped.Extra["mode"])
}

// TestInvoke_NoMatch verifies that Invoke emits nothing when the target
// does not match any probe (returns 404 for all requests).
func TestInvoke_NoMatch(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	origTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	defer func() { http.DefaultTransport = origTransport }()

	host, port := hostPort(t, srv)

	cap := NewCapability()
	input := capmodel.Port{
		Service: "https",
		Port:    port,
		Parent:  capmodel.Asset{DNS: host},
	}

	var emitted []any
	err := cap.Invoke(capability.ExecutionContext{}, input, collectEmitted(&emitted))
	require.NoError(t, err, "Invoke should not error on a non-matching target")
	assert.Empty(t, emitted, "Invoke should emit no models when no probe matches")
}

// TestMatch_ServiceFilter verifies that Match accepts HTTP-based services
// and rejects non-HTTP services. This is the gating function that determines
// whether Julius should even attempt to scan a port.
func TestMatch_ServiceFilter(t *testing.T) {
	cap := NewCapability()
	ctx := capability.ExecutionContext{}

	tests := []struct {
		service string
		wantErr bool
	}{
		{service: "http", wantErr: false},
		{service: "https", wantErr: false},
		{service: "https-alt", wantErr: false},
		{service: "http-proxy", wantErr: false},
		{service: "ssh", wantErr: true},
		{service: "ftp", wantErr: true},
		{service: "", wantErr: true},
		{service: "dns", wantErr: true},
		{service: "smtp", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			input := capmodel.Port{Service: tt.service}
			err := cap.Match(ctx, input)
			if tt.wantErr {
				assert.Error(t, err, "Match should reject service %q", tt.service)
				assert.Contains(t, err.Error(), "julius requires http/https service",
					"error message should explain the rejection")
			} else {
				assert.NoError(t, err, "Match should accept service %q", tt.service)
			}
		})
	}
}
