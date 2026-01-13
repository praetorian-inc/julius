package types

import (
	"fmt"
	"net/http"
	"strings"
)

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
