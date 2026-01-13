package probe

import (
	"io"
	"net/http"
	"os"
	"strings"
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

func TestMatch_StatusCode(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
	rules := types.MatchRules{Status: 200}

	assert.True(t, Match(resp, rules), "Match() should return true for matching status")
}

func TestMatch_StatusCodeMismatch(t *testing.T) {
	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
	rules := types.MatchRules{Status: 200}

	assert.False(t, Match(resp, rules), "Match() should return false for mismatching status")
}

func TestMatch_BodyContains(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"models": []}`)),
		Header:     make(http.Header),
	}
	rules := types.MatchRules{
		Status: 200,
		Body:   types.BodyMatch{Contains: "models"},
	}

	assert.True(t, Match(resp, rules), "Match() should return true for body contains")
}

func TestMatch_HeaderContains(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{"Server": []string{"uvicorn"}},
	}
	rules := types.MatchRules{
		Status: 200,
		Header: types.HeaderMatch{Name: "Server", Contains: "uvicorn"},
	}

	assert.True(t, Match(resp, rules), "Match() should return true for header contains")
}
