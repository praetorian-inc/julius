// pkg/types/types_test.go
package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestProbe_Fields(t *testing.T) {
	p := Probe{
		Type:       "http",
		Path:       "/api/tags",
		Method:     "GET",
		Match:      MatchRules{Status: 200},
		Confidence: "high",
	}

	assert.Equal(t, "/api/tags", p.Path)
	assert.Equal(t, 200, p.Match.Status)
}

func TestMatchRules_BodyMatch(t *testing.T) {
	mr := MatchRules{
		Status: 200,
		Body:   BodyMatch{Contains: "models"},
	}

	assert.Equal(t, "models", mr.Body.Contains)
}

func TestProbe_DefaultMethod(t *testing.T) {
	p := Probe{Path: "/health"}
	p.ApplyDefaults()

	assert.Equal(t, "GET", p.Method)
	assert.Equal(t, "http", p.Type)
	assert.Equal(t, "medium", p.Confidence)
}
