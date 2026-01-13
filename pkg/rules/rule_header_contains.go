package rules

import (
	"fmt"
	"net/http"
	"strings"
)

func init() {
	Register("header.contains", NewHeaderContainsRule)
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
