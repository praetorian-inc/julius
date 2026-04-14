package probe

import (
	"testing"

	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestProbe(name string) *types.Probe {
	return &types.Probe{
		Name: name,
		Requests: []types.Request{
			{Path: "/v1/models"},
			{Path: "/v1/chat/completions"},
		},
	}
}

func makeTestProbeWithModels(name string) *types.Probe {
	return &types.Probe{
		Name: name,
		Requests: []types.Request{
			{Path: "/v1/models"},
		},
		Models: &types.ModelsConfig{
			Path:    "/v1/models",
			Extract: ".data[].id",
		},
	}
}

func makeTestProbeWithAugustus(name string, endpoint string) *types.Probe {
	return &types.Probe{
		Name: name,
		Requests: []types.Request{
			{Path: "/v1/chat/completions"},
		},
		Augustus: &types.AugustusConfig{
			Generator: "openai",
			ConfigTemplate: types.GeneratorConfig{
				Endpoint: endpoint,
			},
		},
	}
}

// TestExpandWithBasePaths_Empty verifies empty base paths returns original probes unchanged.
func TestExpandWithBasePaths_Empty(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a"), makeTestProbe("svc-b")}

	result := ExpandWithBasePaths(probes, nil)

	require.Len(t, result, 2)
	assert.Same(t, probes[0], result[0])
	assert.Same(t, probes[1], result[1])
}

// TestExpandWithBasePaths_EmptySlice verifies an empty slice of base paths returns original probes.
func TestExpandWithBasePaths_EmptySlice(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{})

	require.Len(t, result, 1)
	assert.Same(t, probes[0], result[0])
}

// TestExpandWithBasePaths_SingleBasePath verifies a single base path prepends correctly to request paths.
func TestExpandWithBasePaths_SingleBasePath(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	// Original + 1 expanded copy
	require.Len(t, result, 2)

	// Original unchanged
	assert.Equal(t, "/v1/models", result[0].Requests[0].Path)
	assert.Equal(t, "/v1/chat/completions", result[0].Requests[1].Path)

	// Expanded copy has prepended paths
	assert.Equal(t, "/api/v1/models", result[1].Requests[0].Path)
	assert.Equal(t, "/api/v1/chat/completions", result[1].Requests[1].Path)
}

// TestExpandWithBasePaths_MultipleBasePaths verifies multiple base paths produce correct number of copies.
func TestExpandWithBasePaths_MultipleBasePaths(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a"), makeTestProbe("svc-b")}

	result := ExpandWithBasePaths(probes, []string{"/api", "/proxy"})

	// 2 originals + 2 base paths * 2 probes = 6 total
	require.Len(t, result, 6)

	// Originals first
	assert.Equal(t, "svc-a", result[0].Name)
	assert.Equal(t, "svc-b", result[1].Name)

	// /api expansions
	assert.Equal(t, "/api/v1/models", result[2].Requests[0].Path)
	assert.Equal(t, "/api/v1/models", result[3].Requests[0].Path)

	// /proxy expansions
	assert.Equal(t, "/proxy/v1/models", result[4].Requests[0].Path)
	assert.Equal(t, "/proxy/v1/models", result[5].Requests[0].Path)
}

// TestExpandWithBasePaths_ModelsPath verifies models path gets expanded.
func TestExpandWithBasePaths_ModelsPath(t *testing.T) {
	probes := []*types.Probe{makeTestProbeWithModels("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	require.Len(t, result, 2)

	// Original models path unchanged
	require.NotNil(t, result[0].Models)
	assert.Equal(t, "/v1/models", result[0].Models.Path)

	// Expanded models path has prefix
	require.NotNil(t, result[1].Models)
	assert.Equal(t, "/api/v1/models", result[1].Models.Path)
}

// TestExpandWithBasePaths_NilModels verifies probes without models are handled safely.
func TestExpandWithBasePaths_NilModels(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	require.Len(t, result, 2)
	assert.Nil(t, result[1].Models)
}

// TestExpandWithBasePaths_AugustusTargetEndpoint verifies $TARGET endpoint gets base path appended.
func TestExpandWithBasePaths_AugustusTargetEndpoint(t *testing.T) {
	probes := []*types.Probe{makeTestProbeWithAugustus("svc-a", "$TARGET")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	require.Len(t, result, 2)

	// Original augustus unchanged
	require.NotNil(t, result[0].Augustus)
	assert.Equal(t, "$TARGET", result[0].Augustus.ConfigTemplate.Endpoint)

	// Expanded augustus has base path appended to $TARGET
	require.NotNil(t, result[1].Augustus)
	assert.Equal(t, "$TARGET/api", result[1].Augustus.ConfigTemplate.Endpoint)
}

// TestExpandWithBasePaths_AugustusOtherEndpoint verifies non-$TARGET endpoints are left unchanged.
func TestExpandWithBasePaths_AugustusOtherEndpoint(t *testing.T) {
	probes := []*types.Probe{makeTestProbeWithAugustus("svc-a", "https://custom.endpoint.com")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	require.Len(t, result, 2)
	require.NotNil(t, result[1].Augustus)
	assert.Equal(t, "https://custom.endpoint.com", result[1].Augustus.ConfigTemplate.Endpoint)
}

// TestExpandWithBasePaths_NilAugustus verifies probes without augustus are handled safely.
func TestExpandWithBasePaths_NilAugustus(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"/api"})

	require.Len(t, result, 2)
	assert.Nil(t, result[1].Augustus)
}

// TestExpandWithBasePaths_TrailingSlash verifies trailing slashes on base paths are trimmed.
func TestExpandWithBasePaths_TrailingSlash(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"/api/"})

	require.Len(t, result, 2)
	assert.Equal(t, "/api/v1/models", result[1].Requests[0].Path)
}

// TestExpandWithBasePaths_MissingLeadingSlash verifies base paths without a leading slash get one added.
func TestExpandWithBasePaths_MissingLeadingSlash(t *testing.T) {
	probes := []*types.Probe{makeTestProbe("svc-a")}

	result := ExpandWithBasePaths(probes, []string{"api"})

	require.Len(t, result, 2)
	assert.Equal(t, "/api/v1/models", result[1].Requests[0].Path)
}

// TestExpandWithBasePaths_OriginalNotMutated verifies the original probe slices are not mutated.
func TestExpandWithBasePaths_OriginalNotMutated(t *testing.T) {
	p := makeTestProbe("svc-a")
	origPath0 := p.Requests[0].Path
	origPath1 := p.Requests[1].Path

	_ = ExpandWithBasePaths([]*types.Probe{p}, []string{"/api"})

	assert.Equal(t, origPath0, p.Requests[0].Path)
	assert.Equal(t, origPath1, p.Requests[1].Path)
}
