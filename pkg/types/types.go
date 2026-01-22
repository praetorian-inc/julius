package types

import (
	"fmt"
	"strings"

	"github.com/praetorian-inc/augustus/pkg/generator"
	"github.com/praetorian-inc/julius/pkg/rules"
)

type Result struct {
	Target           string             `json:"target"`
	Service          string             `json:"service"`
	Confidence       string             `json:"confidence"`
	MatchedProbe     string             `json:"matched_probe"`
	Category         string             `json:"category"`
	Models           []string           `json:"models,omitempty"`
	GeneratorConfigs []generator.Config `json:"generator_configs,omitempty"`
	Error            string             `json:"error,omitempty"`
}

type OutputWriter interface {
	Write(results []Result) error
}

// AugustusConfig defines how to scan this service with Augustus.
// ConfigTemplate uses $TARGET and $MODEL placeholders that get resolved at runtime.
type AugustusConfig struct {
	Generator      string           `yaml:"generator"`
	ConfigTemplate generator.Config `yaml:"config_template"`
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

func (p *ProbeDefinition) BuildGeneratorConfigs(target string, models []string) []generator.Config {
	if p.Augustus == nil {
		return nil
	}

	if len(models) == 0 {
		config := resolveGeneratorConfig(p.Augustus.ConfigTemplate, p.Augustus.Generator, target, "")
		return []generator.Config{config}
	}

	configs := make([]generator.Config, 0, len(models))
	for _, model := range models {
		config := resolveGeneratorConfig(p.Augustus.ConfigTemplate, p.Augustus.Generator, target, model)
		configs = append(configs, config)
	}
	return configs
}

func resolveGeneratorConfig(cfg generator.Config, genType, target, model string) generator.Config {
	cfg.Type = genType
	cfg.Endpoint = resolveVars(cfg.Endpoint, target, model)
	cfg.APIKey = resolveVars(cfg.APIKey, target, model)
	cfg.Model = resolveVars(cfg.Model, target, model)
	cfg.Body = resolveVars(cfg.Body, target, model)
	cfg.ResponsePath = resolveVars(cfg.ResponsePath, target, model)
	cfg.Proxy = resolveVars(cfg.Proxy, target, model)

	if cfg.Headers != nil {
		resolved := make(map[string]string, len(cfg.Headers))
		for k, v := range cfg.Headers {
			resolved[k] = resolveVars(v, target, model)
		}
		cfg.Headers = resolved
	}

	return cfg
}

func resolveVars(s, target, model string) string {
	s = strings.ReplaceAll(s, "$TARGET", target)
	if model != "" {
		s = strings.ReplaceAll(s, "$MODEL", model)
	}
	return s
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
