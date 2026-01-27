package rules

import (
	"net/http"
	"strings"
)

type ContentTypeRule struct {
	BaseRule
	Value string
}

func (r *ContentTypeRule) Match(resp *http.Response, body []byte) bool {
	if resp == nil {
		return r.Not
	}

	contentType := resp.Header.Get("Content-Type")
	// Check if content-type contains the expected value (e.g., "application/json" matches "application/json; charset=utf-8")
	matched := strings.Contains(strings.ToLower(contentType), strings.ToLower(r.Value))

	if r.Not {
		return !matched
	}
	return matched
}

func init() {
	Register("content-type", func(raw *RawRule) (Rule, error) {
		val, err := toString(raw.Value)
		if err != nil {
			return nil, err
		}
		return &ContentTypeRule{
			BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
			Value:    val,
		}, nil
	})
}
