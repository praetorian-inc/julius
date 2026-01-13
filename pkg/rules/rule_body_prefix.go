package rules

import (
	"fmt"
	"net/http"
	"strings"
)

func init() {
	Register("body.prefix", NewBodyPrefixRule)
}

type BodyPrefixRule struct {
	BaseRule
	Value string
}

func (r BodyPrefixRule) Match(resp *http.Response, body []byte) bool {
	result := strings.HasPrefix(string(body), r.Value)
	if r.Not {
		return !result
	}
	return result
}

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
