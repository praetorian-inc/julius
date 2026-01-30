package types

import (
	"fmt"
	"strings"

	"github.com/praetorian-inc/julius/pkg/rules"
)

// Specificity constants for common levels
const (
	SpecificityGeneric = 1   // Fallback probes (lowest priority)
	SpecificityLow     = 25  // Broad detection
	SpecificityMedium  = 50  // Default
	SpecificityHigh    = 75  // Service-specific markers
	SpecificityExact   = 100 // Definitive identification
)

type Result struct {
	Target           string            `json:"target"`
	Service          string            `json:"service"`
	MatchedRequest   string            `json:"matched_request"`
	Category         string            `json:"category"`
	Specificity      int               `json:"specificity"`
	Models           []string          `json:"models,omitempty"`
	GeneratorConfigs []GeneratorConfig `json:"generator_configs,omitempty"`
	Error            string            `json:"error,omitempty"`
}

type OutputWriter interface {
	Write(results []Result) error
}

const (
	RequireAny = "any" // Default: match if ANY request succeeds
	RequireAll = "all" // Match only if ALL requests succeed
)

// Probe defines a service detection probe with one or more HTTP requests
type Probe struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Category    string          `yaml:"category"`
	PortHint    int             `yaml:"port_hint"`
	Specificity int             `yaml:"specificity"`       // 1-100, 0 treated as default (50)
	Require     string          `yaml:"require,omitempty"` // "any" (default) or "all"
	APIDocs     string          `yaml:"api_docs"`
	Requests    []Request       `yaml:"requests"`
	Models      *ModelsConfig   `yaml:"models,omitempty"`
	Augustus    *AugustusConfig `yaml:"augustus,omitempty"`
}

func (p *Probe) RequiresAll() bool {
	return strings.ToLower(p.Require) == RequireAll
}

// GetSpecificity returns the probe's specificity, defaulting to 50 if not set
func (p *Probe) GetSpecificity() int {
	if p.Specificity <= 0 {
		return SpecificityMedium // Default 50
	}
	return p.Specificity
}

type ModelsConfig struct {
	Path    string            `yaml:"path"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Extract string            `yaml:"extract"`
}

// Request defines a single HTTP request within a probe
type Request struct {
	Type     string            `yaml:"type"`
	Path     string            `yaml:"path"`
	Method   string            `yaml:"method"`
	Body     string            `yaml:"body,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	RawMatch []rules.RawRule   `yaml:"match"`
}

func (r *Request) ApplyDefaults() {
	if r.Type == "" {
		r.Type = "http"
	}
	if r.Method == "" {
		r.Method = "GET"
	}
}

func (r *Request) GetRules() ([]rules.Rule, error) {
	result := make([]rules.Rule, 0, len(r.RawMatch))
	for i, raw := range r.RawMatch {
		rule, err := raw.ToRule()
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		result = append(result, rule)
	}
	return result, nil
}
