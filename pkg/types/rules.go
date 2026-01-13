// pkg/types/rules.go
package types

import (
	"fmt"
	"net/http"
	"strings"
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

// NewStatusRule creates a StatusRule from a RawRule
func NewStatusRule(raw *RawRule) (Rule, error) {
	val, err := toInt(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("status %w", err)
	}
	return &StatusRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Status:   val,
	}, nil
}

// NewBodyContainsRule creates a BodyContainsRule from a RawRule
func NewBodyContainsRule(raw *RawRule) (Rule, error) {
	val, err := toString(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("body.contains %w", err)
	}
	return &BodyContainsRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Value:    val,
	}, nil
}

// NewBodyPrefixRule creates a BodyPrefixRule from a RawRule
func NewBodyPrefixRule(raw *RawRule) (Rule, error) {
	val, err := toString(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("body.prefix %w", err)
	}
	return &BodyPrefixRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Value:    val,
	}, nil
}

// NewHeaderContainsRule creates a HeaderContainsRule from a RawRule
func NewHeaderContainsRule(raw *RawRule) (Rule, error) {
	val, err := toString(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("header.contains %w", err)
	}
	return &HeaderContainsRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Header:   raw.Header,
		Value:    val,
	}, nil
}

// NewHeaderPrefixRule creates a HeaderPrefixRule from a RawRule
func NewHeaderPrefixRule(raw *RawRule) (Rule, error) {
	val, err := toString(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("header.prefix %w", err)
	}
	return &HeaderPrefixRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Header:   raw.Header,
		Value:    val,
	}, nil
}

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

// StatusRule matches HTTP status codes
type StatusRule struct {
	BaseRule
	Status int
}

// Match checks if the response status code matches the rule
func (s StatusRule) Match(resp *http.Response, body []byte) bool {
	matches := resp.StatusCode == s.Status

	// If Not is true, invert the result
	if s.Not {
		return !matches
	}

	return matches
}

// BodyContainsRule matches if response body contains a string value
type BodyContainsRule struct {
	BaseRule
	Value string
}

// Match checks if the response body contains the specified string
func (r BodyContainsRule) Match(resp *http.Response, body []byte) bool {
	result := strings.Contains(string(body), r.Value)
	if r.Not {
		return !result
	}
	return result
}

// BodyPrefixRule matches if response body starts with a string value
type BodyPrefixRule struct {
	BaseRule
	Value string
}

// Match checks if the response body starts with the specified string
func (r BodyPrefixRule) Match(resp *http.Response, body []byte) bool {
	result := strings.HasPrefix(string(body), r.Value)
	if r.Not {
		return !result
	}
	return result
}

// HeaderContainsRule matches if a header value contains a string
type HeaderContainsRule struct {
	BaseRule
	Header string
	Value  string
}

// Match checks if the header value contains the specified string
func (r HeaderContainsRule) Match(resp *http.Response, body []byte) bool {
	headerVal := resp.Header.Get(r.Header)
	if headerVal == "" {
		if r.Not {
			return true // Header not present, NOT contains = true
		}
		return false
	}
	result := strings.Contains(headerVal, r.Value)
	if r.Not {
		return !result
	}
	return result
}

// HeaderPrefixRule matches if a header value starts with a string
type HeaderPrefixRule struct {
	BaseRule
	Header string
	Value  string
}

// Match checks if the header value starts with the specified string
func (r HeaderPrefixRule) Match(resp *http.Response, body []byte) bool {
	headerVal := resp.Header.Get(r.Header)
	if headerVal == "" {
		if r.Not {
			return true
		}
		return false
	}
	result := strings.HasPrefix(headerVal, r.Value)
	if r.Not {
		return !result
	}
	return result
}
