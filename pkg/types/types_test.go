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
		Target:       "10.0.1.5:11434",
		Service:      "ollama",
		Confidence:   "high",
		MatchedProbe: "/api/tags",
		Category:     "self-hosted",
	}

	data, err := json.Marshal(r)
	require.NoError(t, err, "json.Marshal() should not error")

	var decoded Result
	require.NoError(t, json.Unmarshal(data, &decoded), "json.Unmarshal() should not error")

	assert.Equal(t, r.Target, decoded.Target)
	assert.Equal(t, r.Service, decoded.Service)
}

func TestProbeDefinition_Fields(t *testing.T) {
	pd := ProbeDefinition{
		Name:        "ollama",
		Description: "Ollama local LLM server",
		Category:    "self-hosted",
		PortHint:    11434,
		APIDocs:     "https://example.com",
		Probes:      []Probe{},
	}

	assert.Equal(t, "ollama", pd.Name)
	assert.Equal(t, 11434, pd.PortHint)
}

func TestProbe_DefaultMethod(t *testing.T) {
	p := Probe{Path: "/health"}
	p.ApplyDefaults()

	assert.Equal(t, "GET", p.Method)
	assert.Equal(t, "http", p.Type)
	assert.Equal(t, "medium", p.Confidence)
}

func TestProbe_NewFields(t *testing.T) {
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
	var probe Probe
	err := yaml.Unmarshal([]byte(yamlData), &probe)
	require.NoError(t, err)

	assert.Equal(t, "POST", probe.Method)
	assert.Equal(t, `{"test": "data"}`, probe.Body)
	assert.Equal(t, "application/json", probe.Headers["Content-Type"])
	assert.Equal(t, "Bearer token", probe.Headers["Authorization"])
	assert.Len(t, probe.RawMatch, 2)

	// Test GetRules
	ruleList, err := probe.GetRules()
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
probes: []
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
probes: []
`,
			expectModels: false,
		},
		{
			name: "minimal models config",
			yaml: `
name: test
probes: []
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
			var pd ProbeDefinition
			err := yaml.Unmarshal([]byte(tt.yaml), &pd)
			require.NoError(t, err, "YAML parsing should succeed")

			if tt.expectModels {
				require.NotNil(t, pd.Models, "Models should not be nil")
				assert.Equal(t, tt.expectedPath, pd.Models.Path)
				assert.Equal(t, tt.expectedMethod, pd.Models.Method)
				assert.Equal(t, tt.expectedExtract, pd.Models.Extract)
			} else {
				assert.Nil(t, pd.Models, "Models should be nil when not specified")
			}
		})
	}
}
