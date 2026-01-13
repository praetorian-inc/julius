// pkg/types/rules.go
package types

import (
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
