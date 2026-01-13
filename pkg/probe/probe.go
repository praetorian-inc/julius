package probe

import (
	"embed"
	"fmt"
	"io/fs"
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
	return loadProbesFromFS(os.DirFS(dir), ".")
}

func LoadProbesFromFS(fsys embed.FS, dir string) ([]*types.ProbeDefinition, error) {
	return loadProbesFromFS(fsys, dir)
}

func loadProbesFromFS(fsys fs.FS, dir string) ([]*types.ProbeDefinition, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("reading probe directory: %w", err)
	}

	var probes []*types.ProbeDefinition
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isTemplateFileExt(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := fs.ReadFile(fsys, path)
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

func MatchRules(resp *http.Response, body []byte, ruleList []rules.Rule) bool {
	for _, rule := range ruleList {
		if !rule.Match(resp, body) {
			return false
		}
	}
	return true
}

func isTemplateFileExt(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}
