package types

import "fmt"

type Result struct {
	Target       string `json:"target"`
	Service      string `json:"service"`
	Confidence   string `json:"confidence"`
	MatchedProbe string `json:"matched_probe"`
	Category     string `json:"category"`
}

type OutputWriter interface {
	Write(results []Result) error
}

type ProbeDefinition struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Category    string  `yaml:"category"`
	PortHint    int     `yaml:"port_hint"`
	APIDocs     string  `yaml:"api_docs"`
	Probes      []Probe `yaml:"probes"`
}

type Probe struct {
	Type       string            `yaml:"type"`
	Path       string            `yaml:"path"`
	Method     string            `yaml:"method"`
	Body       string            `yaml:"body,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	RawMatch   []RawRule         `yaml:"match"`
	Confidence string            `yaml:"confidence"`
}

type MatchRules struct {
	Status int         `yaml:"status"`
	Body   BodyMatch   `yaml:"body"`
	Header HeaderMatch `yaml:"header"`
}

type BodyMatch struct {
	Contains string `yaml:"contains"`
	Prefix   string `yaml:"prefix"`
}

type HeaderMatch struct {
	Name     string `yaml:"name"`
	Contains string `yaml:"contains"`
	Prefix   string `yaml:"prefix"`
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

// GetRules converts RawMatch to typed Rule slice
func (p *Probe) GetRules() ([]Rule, error) {
	rules := make([]Rule, 0, len(p.RawMatch))
	for i, raw := range p.RawMatch {
		rule, err := raw.ToRule()
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
