// pkg/scanner/scanner_test.go
package scanner

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScanner(t *testing.T) {
	timeout := 5 * time.Second
	s := NewScanner(timeout)

	require.NotNil(t, s, "NewScanner should not return nil")
	assert.NotNil(t, s.client, "Scanner.client should not be nil")
	assert.Equal(t, timeout, s.client.Timeout)
}

func TestProbe_Match(t *testing.T) {
	// Create test server that returns matching response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response body"))
	}))
	defer server.Close()

	s := NewScanner(5 * time.Second)
	probe := types.Probe{
		Type:   "http",
		Path:   "/",
		Method: "GET",
		Match: types.MatchRules{
			Status: 200,
			Body: types.BodyMatch{
				Contains: "test response",
			},
		},
	}

	matched, err := s.Probe(server.URL, probe)
	require.NoError(t, err, "Probe should not return error")
	assert.True(t, matched, "Probe should return true for matching response")
}

func TestProbe_NoMatch(t *testing.T) {
	// Create test server that returns non-matching response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("different body"))
	}))
	defer server.Close()

	s := NewScanner(5 * time.Second)
	probe := types.Probe{
		Type:   "http",
		Path:   "/",
		Method: "GET",
		Match: types.MatchRules{
			Status: 200,
			Body: types.BodyMatch{
				Contains: "test response",
			},
		},
	}

	matched, err := s.Probe(server.URL, probe)
	require.NoError(t, err, "Probe should not return error")
	assert.False(t, matched, "Probe should return false for non-matching response")
}

func TestScan_FirstMatch(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("openai response"))
	}))
	defer server.Close()

	s := NewScanner(5 * time.Second)
	probes := []*types.ProbeDefinition{
		{
			Name:     "Claude",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/messages",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "claude",
						},
					},
				},
			},
		},
		{
			Name:     "OpenAI",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/chat/completions",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "openai",
						},
					},
				},
			},
		},
	}

	result := s.Scan(server.URL, probes)

	require.NotNil(t, result, "Scan should return result")
	assert.Equal(t, "OpenAI", result.Service)
	assert.Equal(t, "LLM", result.Category)
	assert.Equal(t, server.URL, result.Target)
}

func TestScan_NoMatch(t *testing.T) {
	// Create test server that doesn't match any probe
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown service"))
	}))
	defer server.Close()

	s := NewScanner(5 * time.Second)
	probes := []*types.ProbeDefinition{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/chat/completions",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "openai",
						},
					},
				},
			},
		},
	}

	result := s.Scan(server.URL, probes)

	assert.Nil(t, result, "Scan should return nil when no match")
}

func TestScanAll(t *testing.T) {
	// Create two test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("openai response"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("claude response"))
	}))
	defer server2.Close()

	s := NewScanner(5 * time.Second)
	targets := []string{server1.URL, server2.URL}
	probes := []*types.ProbeDefinition{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/chat/completions",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "openai",
						},
					},
				},
			},
		},
		{
			Name:     "Claude",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/messages",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "claude",
						},
					},
				},
			},
		},
	}

	results := s.ScanAll(targets, probes)

	require.Len(t, results, 2, "ScanAll should return 2 results")

	// Check first result
	assert.Equal(t, "OpenAI", results[0].Service)
	assert.Equal(t, server1.URL, results[0].Target)

	// Check second result
	assert.Equal(t, "Claude", results[1].Service)
	assert.Equal(t, server2.URL, results[1].Target)
}

func TestScanAll_SomeNoMatch(t *testing.T) {
	// Create servers - one matches, one doesn't
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("openai response"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown service"))
	}))
	defer server2.Close()

	s := NewScanner(5 * time.Second)
	targets := []string{server1.URL, server2.URL}
	probes := []*types.ProbeDefinition{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Probes: []types.Probe{
				{
					Path:   "/v1/chat/completions",
					Method: "POST",
					Match: types.MatchRules{
						Body: types.BodyMatch{
							Contains: "openai",
						},
					},
				},
			},
		},
	}

	results := s.ScanAll(targets, probes)

	require.Len(t, results, 1, "ScanAll should return 1 result")
	assert.Equal(t, "OpenAI", results[0].Service)
}

func TestExtractPort(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   int
	}{
		{
			name:   "explicit port",
			target: "http://example.com:8080",
			want:   8080,
		},
		{
			name:   "https explicit port",
			target: "https://example.com:9443",
			want:   9443,
		},
		{
			name:   "http default port",
			target: "http://example.com",
			want:   80,
		},
		{
			name:   "https default port",
			target: "https://example.com",
			want:   443,
		},
		{
			name:   "http with path",
			target: "http://example.com/api/v1",
			want:   80,
		},
		{
			name:   "https with path and port",
			target: "https://example.com:8443/api",
			want:   8443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPort(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}
