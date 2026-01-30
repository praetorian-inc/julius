package types

import (
	"fmt"

	"github.com/praetorian-inc/julius/pkg/rules"
)

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
