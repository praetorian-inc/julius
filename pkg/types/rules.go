package types

import (
	"fmt"
	"net/http"
)

// Rule defines the interface for matching HTTP responses
type Rule interface {
	Match(resp *http.Response, body []byte) bool
	GetType() string
	IsNegated() bool
}

// BaseRule provides common fields for all rule types
type BaseRule struct {
	Type string
	Not  bool
}

// GetType returns the rule type
func (b BaseRule) GetType() string {
	return b.Type
}

// IsNegated returns whether the rule is negated
func (b BaseRule) IsNegated() bool {
	return b.Not
}

// RawRule is used for YAML unmarshaling before converting to specific rule type
type RawRule struct {
	Type   string `yaml:"type"`
	Not    bool   `yaml:"not,omitempty"`
	Value  any    `yaml:"value,omitempty"`
	Header string `yaml:"header,omitempty"`
}

// Decoder converts a RawRule to a typed Rule
type Decoder func(raw *RawRule) (Rule, error)

// ruleDecoders maps type names to constructor functions
var ruleDecoders = map[string]Decoder{
	"status":          NewStatusRule,
	"body.contains":   NewBodyContainsRule,
	"body.prefix":     NewBodyPrefixRule,
	"header.contains": NewHeaderContainsRule,
	"header.prefix":   NewHeaderPrefixRule,
}

// ToRule converts RawRule to the appropriate Rule implementation
func (r *RawRule) ToRule() (Rule, error) {
	decoder, ok := ruleDecoders[r.Type]
	if !ok {
		return nil, fmt.Errorf("unknown rule type: %s", r.Type)
	}
	return decoder(r)
}

// toInt converts any to int (handles float64 from YAML)
func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case uint64:
		return int(val), nil
	default:
		return 0, fmt.Errorf("value must be int, got %T", v)
	}
}

// toString converts any to string
func toString(v any) (string, error) {
	val, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value must be string, got %T", v)
	}
	return val, nil
}
