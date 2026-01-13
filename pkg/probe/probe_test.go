package probe

import (
	"net/http"
	"os"
	"testing"

	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/praetorian-inc/julius/probes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProbeFromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/probes/valid_probe.yaml")
	require.NoError(t, err, "Failed to read test file")

	pd, err := ParseProbe(data)
	require.NoError(t, err, "ParseProbe() should not error")

	assert.Equal(t, "test-service", pd.Name)
	assert.Equal(t, 8080, pd.PortHint)
	assert.Len(t, pd.Probes, 1)
}

func TestLoadProbesFromDir(t *testing.T) {
	loadedProbes, err := LoadProbesFromDir("../../testdata/probes")
	require.NoError(t, err, "LoadProbesFromDir() should not error")

	assert.Len(t, loadedProbes, 1)
	assert.Equal(t, "test-service", loadedProbes[0].Name)
}

func TestLoadProbesFromDir_NotExists(t *testing.T) {
	_, err := LoadProbesFromDir("/nonexistent/path")
	assert.Error(t, err, "LoadProbesFromDir() should error for nonexistent path")
}

func TestLoadProbesFromFS(t *testing.T) {
	loadedProbes, err := LoadProbesFromFS(probes.EmbeddedProbes, ".")
	require.NoError(t, err, "LoadProbesFromFS() should not error")

	assert.GreaterOrEqual(t, len(loadedProbes), 1)
}

func TestSortProbesByPortHint(t *testing.T) {
	probes := []*types.ProbeDefinition{
		{Name: "generic", PortHint: 0},
		{Name: "ollama", PortHint: 11434},
		{Name: "vllm", PortHint: 8000},
	}

	sorted := SortProbesByPortHint(probes, 11434)

	assert.Equal(t, "ollama", sorted[0].Name)
}

func TestSortProbesByPortHint_NoMatch(t *testing.T) {
	probes := []*types.ProbeDefinition{
		{Name: "a", PortHint: 8000},
		{Name: "b", PortHint: 9000},
	}

	sorted := SortProbesByPortHint(probes, 11434)
	// Order should be unchanged
	assert.Len(t, sorted, 2)
}

func TestMatchRules_AllPass(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Server": []string{"uvicorn"}},
	}
	body := []byte(`{"models": []}`)

	rules := []types.Rule{
		&types.StatusRule{BaseRule: types.BaseRule{Type: "status"}, Status: 200},
		&types.BodyContainsRule{BaseRule: types.BaseRule{Type: "body.contains"}, Value: "models"},
	}

	result := MatchRules(resp, body, rules)
	assert.True(t, result)
}

func TestMatchRules_NegationRejects(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
	}
	body := []byte(`<!DOCTYPE html><body>OK</body>`)

	rules := []types.Rule{
		&types.StatusRule{BaseRule: types.BaseRule{Type: "status"}, Status: 200},
		&types.BodyContainsRule{BaseRule: types.BaseRule{Type: "body.contains", Not: true}, Value: "<!DOCTYPE html"},
	}

	result := MatchRules(resp, body, rules)
	assert.False(t, result) // Should fail because body contains HTML (negated rule fails)
}

func TestMatchRules_EmptyRules(t *testing.T) {
	resp := &http.Response{StatusCode: 200}
	result := MatchRules(resp, nil, []types.Rule{})
	assert.True(t, result) // Empty rules should pass (nothing to check)
}
