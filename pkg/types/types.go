package types

import (
	"fmt"

	"github.com/praetorian-inc/julius/pkg/rules"
)

type Result struct {
	Target       string   `json:"target"`
	Service      string   `json:"service"`
	Confidence   string   `json:"confidence"`
	MatchedProbe string   `json:"matched_probe"`
	Category     string   `json:"category"`
	Models []string `json:"models,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type OutputWriter interface {
	Write(results []Result) error
}

type ProbeDefinition struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Category    string        `yaml:"category"`
	PortHint    int           `yaml:"port_hint"`
	APIDocs     string        `yaml:"api_docs"`
	Probes      []Probe       `yaml:"probes"`
	Models      *ModelsConfig `yaml:"models,omitempty"`
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
