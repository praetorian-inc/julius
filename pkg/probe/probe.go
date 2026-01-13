package probe

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/praetorian-inc/julius/pkg/rules"
	"github.com/praetorian-inc/julius/pkg/types"
)

func ParseProbe(data []byte) (*types.ProbeDefinition, error) {
	var pd types.ProbeDefinition
	if err := yaml.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("parsing probe YAML: %w", err)
	}

	for i := range pd.Probes {
		pd.Probes[i].ApplyDefaults()
	}

	return &pd, nil
}

func LoadProbesFromDir(dir string) ([]*types.ProbeDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading probe directory: %w", err)
	}

	var probes []*types.ProbeDefinition
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		pd, err := ParseProbe(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		probes = append(probes, pd)
	}

	return probes, nil
}

func LoadProbesFromFS(fsys embed.FS, dir string) ([]*types.ProbeDefinition, error) {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading embedded probe directory: %w", err)
	}

	var probes []*types.ProbeDefinition
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := fsys.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading embedded %s: %w", path, err)
		}

		pd, err := ParseProbe(data)
		if err != nil {
			return nil, fmt.Errorf("parsing embedded %s: %w", path, err)
		}

		probes = append(probes, pd)
	}

	return probes, nil
}

func SortProbesByPortHint(probes []*types.ProbeDefinition, targetPort int) []*types.ProbeDefinition {
	sorted := make([]*types.ProbeDefinition, len(probes))
	copy(sorted, probes)

	sort.SliceStable(sorted, func(i, j int) bool {
		iMatch := sorted[i].PortHint == targetPort
		jMatch := sorted[j].PortHint == targetPort
		return iMatch && !jMatch
	})

	return sorted
}

// MatchRules checks if all rules match the response
func MatchRules(resp *http.Response, body []byte, ruleList []rules.Rule) bool {
	for _, rule := range ruleList {
		if !rule.Match(resp, body) {
			return false
		}
	}
	return true
}
