package rules

import (
	"fmt"
	"net/http"
)

// StatusRule matches HTTP status codes
type StatusRule struct {
	BaseRule
	Status int
}

// Match checks if the response status code matches the rule
func (r StatusRule) Match(resp *http.Response, body []byte) bool {
	matches := resp.StatusCode == r.Status
	if r.Not {
		return !matches
	}
	return matches
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
