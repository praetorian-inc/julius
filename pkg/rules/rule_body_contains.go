package rules

import (
	"fmt"
	"net/http"
	"strings"
)

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
