package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestResult_JSONMarshal(t *testing.T) {
	r := Result{
		Target:         "10.0.1.5:11434",
		Service:        "ollama",
		Confidence:     "high",
		MatchedRequest: "/api/tags",
		Category:       "self-hosted",
		Specificity:    100,
	}

	data, err := json.Marshal(r)
	require.NoError(t, err, "json.Marshal() should not error")

	var decoded Result
	require.NoError(t, json.Unmarshal(data, &decoded), "json.Unmarshal() should not error")

	assert.Equal(t, r.Target, decoded.Target)
	assert.Equal(t, r.Service, decoded.Service)
	assert.Equal(t, r.Specificity, decoded.Specificity)
}

func TestProbe_Fields(t *testing.T) {
	p := Probe{
		Name:        "ollama",
		Description: "Ollama local LLM server",
		Category:    "self-hosted",
		PortHint:    11434,
		Specificity: 100,
		APIDocs:     "https://example.com",
		Requests:    []Request{},
	}

	assert.Equal(t, "ollama", p.Name)
	assert.Equal(t, 11434, p.PortHint)
	assert.Equal(t, 100, p.GetSpecificity())
}

func TestProbe_GetSpecificity(t *testing.T) {
	tests := []struct {
		name        string
		specificity int
		expected    int
	}{
		{"explicit 100", 100, 100},
		{"explicit 75", 75, 75},
		{"explicit 1 (generic)", 1, 1},
		{"zero defaults to 50", 0, 50},
		{"negative defaults to 50", -1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Probe{Specificity: tt.specificity}
			assert.Equal(t, tt.expected, p.GetSpecificity())
		})
	}
}

func TestRequest_DefaultMethod(t *testing.T) {
	r := Request{Path: "/health"}
	r.ApplyDefaults()

	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "http", r.Type)
	assert.Equal(t, "medium", r.Confidence)
}

func TestRequest_NewFields(t *testing.T) {
	yamlData := `
type: http
path: /api/auth
method: POST
body: '{"test": "data"}'
headers:
  Content-Type: application/json
  Authorization: Bearer token
match:
  - type: status
    value: 200
  - type: body.contains
    value: success
confidence: high
`
	var req Request
	err := yaml.Unmarshal([]byte(yamlData), &req)
	require.NoError(t, err)

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, `{"test": "data"}`, req.Body)
	assert.Equal(t, "application/json", req.Headers["Content-Type"])
	assert.Equal(t, "Bearer token", req.Headers["Authorization"])
	assert.Len(t, req.RawMatch, 2)

	// Test GetRules
	ruleList, err := req.GetRules()
	require.NoError(t, err)
	assert.Len(t, ruleList, 2)
	assert.Equal(t, "status", ruleList[0].GetType())
	assert.Equal(t, "body.contains", ruleList[1].GetType())
}

func TestModelsConfig(t *testing.T) {
	tests := []struct {
		name            string
		yaml            string
		expectModels    bool
		expectedPath    string
		expectedMethod  string
		expectedExtract string
	}{
		{
			name: "full config",
			yaml: `
name: test
requests: []
models:
  path: /v1/models
  method: GET
  headers:
    Authorization: Bearer test
  extract: ".data[].id"
`,
			expectModels:    true,
			expectedPath:    "/v1/models",
			expectedMethod:  "GET",
			expectedExtract: ".data[].id",
		},
		{
			name: "optional models block",
			yaml: `
name: test
requests: []
`,
			expectModels: false,
		},
		{
			name: "minimal models config",
			yaml: `
name: test
requests: []
models:
  path: /api/tags
  extract: ".models[].name"
`,
			expectModels:    true,
			expectedPath:    "/api/tags",
			expectedMethod:  "",
			expectedExtract: ".models[].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Probe
			err := yaml.Unmarshal([]byte(tt.yaml), &p)
			require.NoError(t, err, "YAML parsing should succeed")

			if tt.expectModels {
				require.NotNil(t, p.Models, "Models should not be nil")
				assert.Equal(t, tt.expectedPath, p.Models.Path)
				assert.Equal(t, tt.expectedMethod, p.Models.Method)
				assert.Equal(t, tt.expectedExtract, p.Models.Extract)
			} else {
				assert.Nil(t, p.Models, "Models should be nil when not specified")
			}
		})
	}
}

func TestResultFields(t *testing.T) {
	tests := []struct {
		name           string
		models         []string
		err            string
		expectedModels int
		expectError    bool
	}{
		{
			name:           "with models",
			models:         []string{"gpt-4", "gpt-3.5-turbo"},
			err:            "",
			expectedModels: 2,
			expectError:    false,
		},
		{
			name:           "with error",
			models:         []string{},
			err:            "models request returned 401",
			expectedModels: 0,
			expectError:    true,
		},
		{
			name:           "empty fields",
			models:         []string{},
			err:            "",
			expectedModels: 0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Result{
				Target:      "https://example.com",
				Service:     "openai",
				Confidence:  "high",
				Specificity: 50,
				Models:      tt.models,
				Error:       tt.err,
			}

			assert.Len(t, result.Models, tt.expectedModels)
			if tt.expectError {
				assert.NotEmpty(t, result.Error)
			} else {
				assert.Empty(t, result.Error)
			}
		})
	}
}

func TestSpecificityConstants(t *testing.T) {
	assert.Equal(t, 1, SpecificityGeneric)
	assert.Equal(t, 25, SpecificityLow)
	assert.Equal(t, 50, SpecificityMedium)
	assert.Equal(t, 75, SpecificityHigh)
	assert.Equal(t, 100, SpecificityExact)
}
