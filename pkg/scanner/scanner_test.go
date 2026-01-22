// pkg/scanner/scanner_test.go
package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/praetorian-inc/julius/pkg/rules"
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
		RawMatch: []rules.RawRule{
			{Type: "status", Value: 200},
			{Type: "body.contains", Value: "test response"},
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
		RawMatch: []rules.RawRule{
			{Type: "status", Value: 200},
			{Type: "body.contains", Value: "test response"},
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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "claude"},
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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "openai"},
					},
				},
			},
		},
	}

	result := s.Scan(server.URL, probes, false)

	require.NotNil(t, result, "Scan should return result")
	assert.Equal(t, "OpenAI", result.Service)
	assert.Equal(t, "LLM", result.Category)
	assert.Equal(t, server.URL+"/v1/chat/completions", result.Target)
	assert.Equal(t, "/v1/chat/completions", result.MatchedProbe)
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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "openai"},
					},
				},
			},
		},
	}

	result := s.Scan(server.URL, probes, false)

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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "openai"},
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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "claude"},
					},
				},
			},
		},
	}

	results := s.ScanAll(targets, probes, false)

	require.Len(t, results, 2, "ScanAll should return 2 results")

	// Check first result
	assert.Equal(t, "OpenAI", results[0].Service)
	assert.Equal(t, server1.URL+"/v1/chat/completions", results[0].Target)
	assert.Equal(t, "/v1/chat/completions", results[0].MatchedProbe)

	// Check second result
	assert.Equal(t, "Claude", results[1].Service)
	assert.Equal(t, server2.URL+"/v1/messages", results[1].Target)
	assert.Equal(t, "/v1/messages", results[1].MatchedProbe)
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
					RawMatch: []rules.RawRule{
						{Type: "body.contains", Value: "openai"},
					},
				},
			},
		},
	}

	results := s.ScanAll(targets, probes, false)

	require.Len(t, results, 1, "ScanAll should return 1 result")
	assert.Equal(t, "OpenAI", results[0].Service)
}

func TestProbe_WithBodyAndHeaders(t *testing.T) {
	// Create test server that echoes back request body and headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		contentType := r.Header.Get("Content-Type")
		auth := r.Header.Get("Authorization")

		response := fmt.Sprintf(`{"body":"%s","content_type":"%s","auth":"%s"}`,
			string(body), contentType, auth)
		w.WriteHeader(200)
		w.Write([]byte(response))
	}))
	defer server.Close()

	s := NewScanner(5 * time.Second)
	p := types.Probe{
		Path:   "/test",
		Method: "POST",
		Body:   `{"test":"data"}`,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		RawMatch: []rules.RawRule{
			{Type: "status", Value: 200},
			{Type: "body.contains", Value: "test"},
		},
	}

	matched, err := s.Probe(server.URL, p)
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestScanWithModels(t *testing.T) {
	tests := []struct {
		name           string
		fingerprint    string
		modelsResponse string
		modelsStatus   int
		expectModels   []string
		expectError    bool
	}{
		{
			name:           "successful models extraction",
			fingerprint:    `{"models":[{"name":"llama3.2:1b"}]}`,
			modelsResponse: `{"models":[{"name":"llama3.2:1b"}]}`,
			modelsStatus:   http.StatusOK,
			expectModels:   []string{"llama3.2:1b"},
			expectError:    false,
		},
		{
			name:           "models fetch fails",
			fingerprint:    `{"models":[]}`,
			modelsResponse: "",
			modelsStatus:   http.StatusUnauthorized,
			expectModels:   nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/tags":
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.fingerprint))
				case "/api/models":
					w.WriteHeader(tt.modelsStatus)
					if tt.modelsResponse != "" {
						w.Write([]byte(tt.modelsResponse))
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			scanner := NewScanner(5 * time.Second)
			probes := []*types.ProbeDefinition{
				{
					Name:     "ollama",
					Category: "self-hosted",
					Probes: []types.Probe{
						{
							Type:   "http",
							Path:   "/api/tags",
							Method: "GET",
							RawMatch: []rules.RawRule{
								{Type: "status", Value: 200},
							},
							Confidence: "high",
						},
					},
					Models: &types.ModelsConfig{
						Path:    "/api/models",
						Method:  "GET",
						Extract: ".models[].name",
					},
				},
			}

			result := scanner.Scan(server.URL, probes, false)
			require.NotNil(t, result, "expected result")

			assert.Equal(t, "ollama", result.Service)
			assert.Equal(t, tt.expectModels, result.Models)

			if tt.expectError {
				assert.NotEmpty(t, result.Error, "expected error")
			} else {
				assert.Empty(t, result.Error, "expected no error")
			}
		})
	}
}

func TestScanWithoutModelsConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	scanner := NewScanner(5 * time.Second)
	probes := []*types.ProbeDefinition{
		{
			Name:     "test-service",
			Category: "test",
			Probes: []types.Probe{
				{
					Type:   "http",
					Path:   "/health",
					Method: "GET",
					RawMatch: []rules.RawRule{
						{Type: "status", Value: 200},
					},
					Confidence: "medium",
				},
			},
			// No Models config
		},
	}

	result := scanner.Scan(server.URL, probes, false)
	require.NotNil(t, result)

	assert.Equal(t, "test-service", result.Service)
	assert.Empty(t, result.Models)
	assert.Empty(t, result.Error)
}

func TestFetchModels(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		config         *types.ModelsConfig
		expected       []string
		wantErr        bool
	}{
		{
			name:           "successful fetch",
			serverResponse: `{"data":[{"id":"gpt-4"},{"id":"gpt-3.5-turbo"}]}`,
			serverStatus:   http.StatusOK,
			config: &types.ModelsConfig{
				Path:    "/v1/models",
				Method:  "GET",
				Extract: ".data[].id",
			},
			expected: []string{"gpt-4", "gpt-3.5-turbo"},
		},
		{
			name:           "default GET method",
			serverResponse: `{"models":[{"name":"llama"}]}`,
			serverStatus:   http.StatusOK,
			config: &types.ModelsConfig{
				Path:    "/api/tags",
				Extract: ".models[].name",
			},
			expected: []string{"llama"},
		},
		{
			name:         "unauthorized error",
			serverStatus: http.StatusUnauthorized,
			config: &types.ModelsConfig{
				Path:    "/v1/models",
				Extract: ".data[].id",
			},
			wantErr: true,
		},
		{
			name:         "not found error",
			serverStatus: http.StatusNotFound,
			config: &types.ModelsConfig{
				Path:    "/v1/models",
				Extract: ".data[].id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			scanner := NewScanner(5 * time.Second)
			models, err := scanner.fetchModels(server.URL, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, models)
		})
	}
}

func TestFetchModelsWithHeaders(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"model-1"}]}`))
	}))
	defer server.Close()

	scanner := NewScanner(5 * time.Second)
	cfg := &types.ModelsConfig{
		Path:    "/v1/models",
		Method:  "GET",
		Headers: map[string]string{"Authorization": "Bearer test-token"},
		Extract: ".data[].id",
	}

	_, err := scanner.fetchModels(server.URL, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token", receivedAuth)
}

func TestExtractModels(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expr     string
		expected []string
		wantErr  bool
	}{
		{
			name:     "OpenAI format",
			body:     `{"data":[{"id":"gpt-4"},{"id":"gpt-3.5-turbo"}]}`,
			expr:     ".data[].id",
			expected: []string{"gpt-4", "gpt-3.5-turbo"},
		},
		{
			name:     "Ollama format",
			body:     `{"models":[{"name":"llama3.2:1b"},{"name":"mistral:7b"}]}`,
			expr:     ".models[].name",
			expected: []string{"llama3.2:1b", "mistral:7b"},
		},
		{
			name:     "simple array",
			body:     `["model-a","model-b"]`,
			expr:     ".[]",
			expected: []string{"model-a", "model-b"},
		},
		{
			name:     "empty result",
			body:     `{"data":[]}`,
			expr:     ".data[].id",
			expected: []string{},
		},
		{
			name:    "invalid JSON",
			body:    `not json`,
			expr:    ".data[].id",
			wantErr: true,
		},
		{
			name:    "invalid jq expression",
			body:    `{"data":[]}`,
			expr:    ".[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models, err := extractModels([]byte(tt.body), tt.expr)

			if tt.wantErr {
				assert.Error(t, err, "expected error")
				return
			}

			require.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expected, models)
		})
	}
}

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "already normalized",
			target: "https://example.com",
			want:   "https://example.com",
		},
		{
			name:   "trailing slash",
			target: "https://example.com/",
			want:   "https://example.com",
		},
		{
			name:   "multiple trailing slashes",
			target: "https://example.com///",
			want:   "https://example.com",
		},
		{
			name:   "trailing slash with path",
			target: "https://example.com/api/v1/",
			want:   "https://example.com/api/v1",
		},
		{
			name:   "leading whitespace",
			target: "  https://example.com",
			want:   "https://example.com",
		},
		{
			name:   "trailing whitespace",
			target: "https://example.com  ",
			want:   "https://example.com",
		},
		{
			name:   "whitespace and trailing slash",
			target: "  https://example.com/  ",
			want:   "https://example.com",
		},
		{
			name:   "http scheme preserved",
			target: "http://example.com/",
			want:   "http://example.com",
		},
		{
			name:   "no scheme adds https",
			target: "example.com",
			want:   "https://example.com",
		},
		{
			name:   "no scheme with port",
			target: "example.com:8080",
			want:   "https://example.com:8080",
		},
		{
			name:   "no scheme with path and trailing slash",
			target: "example.com/api/",
			want:   "https://example.com/api",
		},
		{
			name:   "empty string",
			target: "",
			want:   "",
		},
		{
			name:   "whitespace only",
			target: "   ",
			want:   "",
		},
		{
			name:   "with port",
			target: "https://example.com:11434/",
			want:   "https://example.com:11434",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTarget(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		want    []string
	}{
		{
			name:    "mixed normalization needs",
			targets: []string{"https://a.com/", "  https://b.com  ", "c.com"},
			want:    []string{"https://a.com", "https://b.com", "https://c.com"},
		},
		{
			name:    "filters empty entries",
			targets: []string{"https://a.com", "", "  ", "https://b.com"},
			want:    []string{"https://a.com", "https://b.com"},
		},
		{
			name:    "empty input",
			targets: []string{},
			want:    nil,
		},
		{
			name:    "all empty entries",
			targets: []string{"", "  ", ""},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTargets(tt.targets)
			assert.Equal(t, tt.want, got)
		})
	}
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
