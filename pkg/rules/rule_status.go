package rules

import (
	"fmt"
	"net/http"
)

func init() {
	Register("status", NewStatusRule)
}

type StatusRule struct {
	BaseRule
	Status int
}

func (r StatusRule) Match(resp *http.Response, body []byte) bool {
	matches := resp.StatusCode == r.Status
	if r.Not {
		return !matches
	}
	return matches
}

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
