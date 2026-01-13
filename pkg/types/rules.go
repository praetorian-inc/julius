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
