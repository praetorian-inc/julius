// pkg/scanner/scanner_test.go
package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/praetorian-inc/julius/pkg/rules"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScanner(t *testing.T) {
	timeout := 5 * time.Second
	s := NewScanner(timeout, 10, 10*1024*1024, nil)

	require.NotNil(t, s, "NewScanner should not return nil")
	assert.NotNil(t, s.client, "Scanner.client should not be nil")
	assert.Equal(t, timeout, s.client.Timeout)
	assert.Equal(t, 10, s.concurrency)
}

func TestNewScanner_DefaultConcurrency(t *testing.T) {
	s := NewScanner(5*time.Second, 0, 10*1024*1024, nil)
	assert.Equal(t, 10, s.concurrency, "should default to 10 concurrency")

	s2 := NewScanner(5*time.Second, -1, 10*1024*1024, nil)
	assert.Equal(t, 10, s2.concurrency, "should default to 10 for negative values")
}

func TestDoRequest_Match(t *testing.T) {
	// Create test server that returns matching response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response body"))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	req := types.Request{
		Type:   "http",
		Path:   "/",
		Method: "GET",
		RawMatch: []rules.RawRule{
			{Type: "status", Value: 200},
			{Type: "body.contains", Value: "test response"},
		},
	}

	matched, err := s.DoRequest(server.URL, req)
	require.NoError(t, err, "DoRequest should not return error")
	assert.True(t, matched, "DoRequest should return true for matching response")
}

func TestDoRequest_NoMatch(t *testing.T) {
	// Create test server that returns non-matching response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("different body"))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	req := types.Request{
		Type:   "http",
		Path:   "/",
		Method: "GET",
		RawMatch: []rules.RawRule{
			{Type: "status", Value: 200},
			{Type: "body.contains", Value: "test response"},
		},
	}

	matched, err := s.DoRequest(server.URL, req)
	require.NoError(t, err, "DoRequest should not return error")
	assert.False(t, matched, "DoRequest should return false for non-matching response")
}

func TestScan_ReturnsAllMatches(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:        "specific-service",
			Category:    "LLM",
			Specificity: 75,
			Requests: []types.Request{
				{
					Path:   "/v1/models",
					Method: "GET",
					RawMatch: []rules.RawRule{
						{Type: "status", Value: 200},
					},
				},
			},
		},
		{
			Name:        "generic-service",
			Category:    "generic",
			Specificity: 1,
			Requests: []types.Request{
				{
					Path:   "/v1/models",
					Method: "GET",
					RawMatch: []rules.RawRule{
						{Type: "status", Value: 200},
					},
				},
			},
		},
	}

	results := s.Scan(server.URL, probes, false)

	require.Len(t, results, 2, "Scan should return all matching probes")
	// Results should be sorted by specificity (highest first)
	assert.Equal(t, "specific-service", results[0].Service)
	assert.Equal(t, 75, results[0].Specificity)
	assert.Equal(t, "generic-service", results[1].Service)
	assert.Equal(t, 1, results[1].Specificity)
}

func TestScan_SortsBySpecificity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{Name: "low", Specificity: 25, Requests: []types.Request{{Path: "/", RawMatch: []rules.RawRule{{Type: "status", Value: 200}}}}},
		{Name: "high", Specificity: 100, Requests: []types.Request{{Path: "/", RawMatch: []rules.RawRule{{Type: "status", Value: 200}}}}},
		{Name: "medium", Specificity: 50, Requests: []types.Request{{Path: "/", RawMatch: []rules.RawRule{{Type: "status", Value: 200}}}}},
		{Name: "generic", Specificity: 1, Requests: []types.Request{{Path: "/", RawMatch: []rules.RawRule{{Type: "status", Value: 200}}}}},
	}

	results := s.Scan(server.URL, probes, false)

	require.Len(t, results, 4)
	assert.Equal(t, "high", results[0].Service)
	assert.Equal(t, "medium", results[1].Service)
	assert.Equal(t, "low", results[2].Service)
	assert.Equal(t, "generic", results[3].Service)
}

func TestScan_NoMatch(t *testing.T) {
	// Create test server that doesn't match any probe
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown service"))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Requests: []types.Request{
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

	results := s.Scan(server.URL, probes, false)

	assert.Empty(t, results, "Scan should return empty slice when no match")
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

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	targets := []string{server1.URL, server2.URL}
	probes := []*types.Probe{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Requests: []types.Request{
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
			Requests: []types.Request{
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

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	targets := []string{server1.URL, server2.URL}
	probes := []*types.Probe{
		{
			Name:     "OpenAI",
			Category: "LLM",
			Requests: []types.Request{
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

func TestDoRequest_WithBodyAndHeaders(t *testing.T) {
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

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	req := types.Request{
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

	matched, err := s.DoRequest(server.URL, req)
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

			scanner := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
			probes := []*types.Probe{
				{
					Name:     "ollama",
					Category: "self-hosted",
					Requests: []types.Request{
						{
							Type:   "http",
							Path:   "/api/tags",
							Method: "GET",
							RawMatch: []rules.RawRule{
								{Type: "status", Value: 200},
							},
						},
					},
					Models: &types.ModelsConfig{
						Path:    "/api/models",
						Method:  "GET",
						Extract: ".models[].name",
					},
				},
			}

			results := scanner.Scan(server.URL, probes, false)
			require.Len(t, results, 1, "expected 1 result")

			assert.Equal(t, "ollama", results[0].Service)
			assert.Equal(t, tt.expectModels, results[0].Models)

			if tt.expectError {
				assert.NotEmpty(t, results[0].Error, "expected error")
			} else {
				assert.Empty(t, results[0].Error, "expected no error")
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

	scanner := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:     "test-service",
			Category: "test",
			Requests: []types.Request{
				{
					Type:   "http",
					Path:   "/health",
					Method: "GET",
					RawMatch: []rules.RawRule{
						{Type: "status", Value: 200},
					},
				},
			},
			// No Models config
		},
	}

	results := scanner.Scan(server.URL, probes, false)
	require.Len(t, results, 1)

	assert.Equal(t, "test-service", results[0].Service)
	assert.Empty(t, results[0].Models)
	assert.Empty(t, results[0].Error)
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

			scanner := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
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

	scanner := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
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

// ============================================================================
// Singleflight and Caching Tests
// ============================================================================

func TestSingleflightDeduplication(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		time.Sleep(100 * time.Millisecond) // Simulate latency
		w.WriteHeader(200)
		w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)

	// Simulate multiple goroutines hitting the same endpoint concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.doHTTPRequest(server.URL, "GET", "/v1/models", "", nil)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), requestCount.Load(), "concurrent requests to same URL should be deduplicated")
}

func TestCachePersistsAcrossCalls(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)

	// First call
	s.doHTTPRequest(server.URL, "GET", "/v1/models", "", nil)
	// Second call (should hit cache)
	s.doHTTPRequest(server.URL, "GET", "/v1/models", "", nil)

	assert.Equal(t, int32(1), requestCount.Load(), "second call should use cache")
}

func TestDifferentURLsNotDeduplicated(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)

	// Different paths = different cache keys
	s.doHTTPRequest(server.URL, "GET", "/v1/models", "", nil)
	s.doHTTPRequest(server.URL, "GET", "/v1/chat", "", nil)

	assert.Equal(t, int32(2), requestCount.Load(), "different URLs should not be deduplicated")
}

func TestConcurrentProbeExecution(t *testing.T) {
	var requestTimes []time.Time
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// 5 probes hitting different endpoints
	probes := make([]*types.Probe, 5)
	for i := 0; i < 5; i++ {
		probes[i] = &types.Probe{
			Name:        fmt.Sprintf("probe-%d", i),
			Specificity: 50,
			Requests: []types.Request{{
				Path:     fmt.Sprintf("/endpoint-%d", i),
				Method:   "GET",
				RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
			}},
		}
	}

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	start := time.Now()
	s.Scan(server.URL, probes, false)
	elapsed := time.Since(start)

	// With concurrency, 5 requests @ 50ms each should take ~50-100ms, not 250ms
	assert.Less(t, elapsed, 200*time.Millisecond, "probes should run concurrently")
}

func TestConcurrencyLimit(t *testing.T) {
	var concurrentCount atomic.Int32
	var maxConcurrent atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := concurrentCount.Add(1)
		for {
			old := maxConcurrent.Load()
			if current <= old || maxConcurrent.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		concurrentCount.Add(-1)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Create more probes than the concurrency limit
	probes := make([]*types.Probe, 20)
	for i := 0; i < 20; i++ {
		probes[i] = &types.Probe{
			Name:        fmt.Sprintf("probe-%d", i),
			Specificity: 50,
			Requests: []types.Request{{
				Path:     fmt.Sprintf("/endpoint-%d", i),
				Method:   "GET",
				RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
			}},
		}
	}

	// Set concurrency limit to 5
	s := NewScanner(5*time.Second, 5, 10*1024*1024, nil)
	s.Scan(server.URL, probes, false)

	assert.LessOrEqual(t, maxConcurrent.Load(), int32(5), "should not exceed concurrency limit")
}

func TestCacheKeyIncludesMethod(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)

	// Same URL, different methods = different cache keys
	s.doHTTPRequest(server.URL, "GET", "/v1/models", "", nil)
	s.doHTTPRequest(server.URL, "POST", "/v1/models", "", nil)

	assert.Equal(t, int32(2), requestCount.Load(), "different methods should not be cached together")
}

func TestCacheKeyIncludesBody(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)

	// Same URL and method, different body = different cache keys
	s.doHTTPRequest(server.URL, "POST", "/v1/chat", `{"a":1}`, nil)
	s.doHTTPRequest(server.URL, "POST", "/v1/chat", `{"b":2}`, nil)

	assert.Equal(t, int32(2), requestCount.Load(), "different bodies should not be cached together")
}

// ============================================================================
// Require All Tests
// ============================================================================

func TestScan_RequireAll_AllMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.Header().Set("Server", "uvicorn")
			w.WriteHeader(200)
			w.Write([]byte(`{"object":"list","data":[]}`))
		case "/tokenize":
			w.WriteHeader(200)
			w.Write([]byte(`{"tokens":[1,2,3]}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:    "test-all",
			Require: "all",
			Requests: []types.Request{
				{
					Path:     "/v1/models",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
				{
					Path:     "/tokenize",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
			},
		},
	}

	results := s.Scan(server.URL, probes, false)

	require.Len(t, results, 1, "should match when all requests succeed")
	assert.Equal(t, "test-all", results[0].Service)
	assert.Equal(t, "/v1/models", results[0].MatchedRequest) // First request path
}

func TestScan_RequireAll_SomeFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(200)
			w.Write([]byte(`{"object":"list","data":[]}`))
		case "/tokenize":
			w.WriteHeader(404) // This one fails
			w.Write([]byte(`{"error":"not found"}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:    "test-all",
			Require: "all",
			Requests: []types.Request{
				{
					Path:     "/v1/models",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
				{
					Path:     "/tokenize",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
			},
		},
	}

	results := s.Scan(server.URL, probes, false)

	assert.Empty(t, results, "should not match when any request fails")
}

func TestScan_RequireAny_Default(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(404)
		case "/tokenize":
			w.WriteHeader(200)
			w.Write([]byte(`{"tokens":[1,2,3]}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name: "test-any", // No require field = default "any"
			Requests: []types.Request{
				{
					Path:     "/v1/models",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
				{
					Path:     "/tokenize",
					Method:   "GET",
					RawMatch: []rules.RawRule{{Type: "status", Value: 200}},
				},
			},
		},
	}

	results := s.Scan(server.URL, probes, false)

	require.Len(t, results, 1, "should match when any request succeeds")
	assert.Equal(t, "test-any", results[0].Service)
	assert.Equal(t, "/tokenize", results[0].MatchedRequest) // Second request matched
}

func TestScan_RequireAll_EmptyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	s := NewScanner(5*time.Second, 10, 10*1024*1024, nil)
	probes := []*types.Probe{
		{
			Name:     "test-empty",
			Require:  "all",
			Requests: []types.Request{}, // Empty
		},
	}

	results := s.Scan(server.URL, probes, false)

	assert.Empty(t, results, "should not match with empty requests")
}

func TestProbe_RequiresAll(t *testing.T) {
	tests := []struct {
		require  string
		expected bool
	}{
		{"all", true},
		{"ALL", true},
		{"All", true},
		{"any", false},
		{"ANY", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.require, func(t *testing.T) {
			p := types.Probe{Require: tt.require}
			assert.Equal(t, tt.expected, p.RequiresAll())
		})
	}
}
