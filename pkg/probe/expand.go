package probe

import (
	"strings"

	"github.com/praetorian-inc/julius/pkg/types"
)

// ExpandWithBasePaths returns a new slice containing the original probes followed by
// copies of each probe with request paths and model/augustus paths prefixed by each
// base path. If basePaths is empty, the original slice is returned unchanged.
func ExpandWithBasePaths(probes []*types.Probe, basePaths []string) []*types.Probe {
	if len(basePaths) == 0 {
		return probes
	}

	result := make([]*types.Probe, len(probes), len(probes)*(1+len(basePaths)))
	copy(result, probes)

	for _, rawBase := range basePaths {
		prefix := normalizeBasePath(rawBase)
		for _, p := range probes {
			result = append(result, cloneProbeWithPrefix(p, prefix))
		}
	}

	return result
}

// normalizeBasePath trims trailing slashes and ensures the path starts with /.
func normalizeBasePath(base string) string {
	base = strings.TrimRight(base, "/")
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	return base
}

// cloneProbeWithPrefix creates a shallow-enough copy of the probe with the given
// prefix prepended to all request paths, the models path (if set), and the
// Augustus endpoint when it equals the $TARGET placeholder.
func cloneProbeWithPrefix(p *types.Probe, prefix string) *types.Probe {
	clone := *p

	// Copy and update request paths.
	clone.Requests = make([]types.Request, len(p.Requests))
	for i, req := range p.Requests {
		req.Path = prefix + req.Path
		clone.Requests[i] = req
	}

	// Copy and update models path if present.
	if p.Models != nil {
		m := *p.Models
		m.Path = prefix + m.Path
		clone.Models = &m
	}

	// Copy and update augustus config if present.
	if p.Augustus != nil {
		aug := *p.Augustus
		cfg := aug.ConfigTemplate
		if cfg.Endpoint == "$TARGET" {
			cfg.Endpoint = "$TARGET" + prefix
		}
		aug.ConfigTemplate = cfg
		clone.Augustus = &aug
	}

	return &clone
}
