package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/praetorian-inc/julius/pkg/rules"
)

// GeneratorConfig contains everything needed to run Augustus against an LLM endpoint.
type GeneratorConfig struct {
	Generator string         `json:"generator"`
	Config    map[string]any `json:"config"`
}

type Result struct {
	Target           string            `json:"target"`
	Service          string            `json:"service"`
	Confidence       string            `json:"confidence"`
	MatchedProbe     string            `json:"matched_probe"`
	Category         string            `json:"category"`
	Models           []string          `json:"models,omitempty"`
	GeneratorConfigs []GeneratorConfig `json:"generator_configs,omitempty"`
	Error            string            `json:"error,omitempty"`
}

type OutputWriter interface {
	Write(results []Result) error
}

// AugustusConfig defines how to scan this service with Augustus
type AugustusConfig struct {
	Generator      string         `yaml:"generator"`
	ConfigTemplate map[string]any `yaml:"config_template"`
}

type ProbeDefinition struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Category    string          `yaml:"category"`
	PortHint    int             `yaml:"port_hint"`
	APIDocs     string          `yaml:"api_docs"`
	Probes      []Probe         `yaml:"probes"`
	Models      *ModelsConfig   `yaml:"models,omitempty"`
	Augustus    *AugustusConfig `yaml:"augustus,omitempty"`
}

// BuildGeneratorConfigs creates Augustus GeneratorConfigs from the template.
// It resolves $TARGET and $MODEL variables, leaving $INPUT for Augustus to handle.
func (p *ProbeDefinition) BuildGeneratorConfigs(target string, models []string) []GeneratorConfig {
	if p.Augustus == nil {
		return nil
	}

	// If no models discovered, emit single config (remove $MODEL references)
	if len(models) == 0 {
		config := resolveConfigTemplate(p.Augustus.ConfigTemplate, target, "")
		return []GeneratorConfig{{
			Generator: p.Augustus.Generator,
			Config:    config,
		}}
	}

	// One config per model
	configs := make([]GeneratorConfig, 0, len(models))
	for _, model := range models {
		config := resolveConfigTemplate(p.Augustus.ConfigTemplate, target, model)
		configs = append(configs, GeneratorConfig{
			Generator: p.Augustus.Generator,
			Config:    config,
		})
	}
	return configs
}

// resolveConfigTemplate deep copies the template and resolves $TARGET and $MODEL variables.
// $INPUT is left as-is for Augustus/generator to resolve at probe time.
func resolveConfigTemplate(template map[string]any, target, model string) map[string]any {
	// Deep copy via JSON round-trip
	data, _ := json.Marshal(template)
	var result map[string]any
	json.Unmarshal(data, &result)

	// Resolve variables in all string values
	resolveVariables(result, target, model)
	return result
}

func resolveVariables(m map[string]any, target, model string) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			resolved := strings.ReplaceAll(val, "$TARGET", target)
			if model != "" {
				resolved = strings.ReplaceAll(resolved, "$MODEL", model)
			}
			m[k] = resolved
		case map[string]any:
			resolveVariables(val, target, model)
		case []any:
			for i, item := range val {
				if str, ok := item.(string); ok {
					resolved := strings.ReplaceAll(str, "$TARGET", target)
					if model != "" {
						resolved = strings.ReplaceAll(resolved, "$MODEL", model)
					}
					val[i] = resolved
				} else if nested, ok := item.(map[string]any); ok {
					resolveVariables(nested, target, model)
				}
			}
		}
	}
}

type ModelsConfig struct {
	Path    string            `yaml:"path"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Extract string            `yaml:"extract"`
}

type Probe struct {
	Type       string            `yaml:"type"`
	Path       string            `yaml:"path"`
	Method     string            `yaml:"method"`
	Body       string            `yaml:"body,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	RawMatch   []rules.RawRule   `yaml:"match"`
	Confidence string            `yaml:"confidence"`
}

func (p *Probe) ApplyDefaults() {
	if p.Type == "" {
		p.Type = "http"
	}
	if p.Method == "" {
		p.Method = "GET"
	}
	if p.Confidence == "" {
		p.Confidence = "medium"
	}
}

func (p *Probe) GetRules() ([]rules.Rule, error) {
	result := make([]rules.Rule, 0, len(p.RawMatch))
	for i, raw := range p.RawMatch {
		rule, err := raw.ToRule()
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		result = append(result, rule)
	}
	return result, nil
}
