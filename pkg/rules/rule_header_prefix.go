package rules

import (
	"fmt"
	"net/http"
	"strings"
)

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
