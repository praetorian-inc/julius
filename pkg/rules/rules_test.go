// pkg/types/rules_test.go
package rules

import (
	"net/http"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusRule_Match(t *testing.T) {
	rule := StatusRule{
		BaseRule: BaseRule{
			Type: "status",
			Not:  false,
		},
		Status: 200,
	}

	resp := &http.Response{
		StatusCode: 200,
	}

	assert.True(t, rule.Match(resp, nil), "StatusRule should match when status code is 200")

	resp.StatusCode = 404
	assert.False(t, rule.Match(resp, nil), "StatusRule should not match when status code is 404")
}

func TestStatusRule_Match_WithNot(t *testing.T) {
	rule := StatusRule{
		BaseRule: BaseRule{
			Type: "status",
			Not:  true, // Negated rule
		},
		Status: 404,
	}

	resp := &http.Response{
		StatusCode: 200,
	}

	assert.True(t, rule.Match(resp, nil), "StatusRule with Not=true should match when status is NOT 404")

	resp.StatusCode = 404
	assert.False(t, rule.Match(resp, nil), "StatusRule with Not=true should not match when status IS 404")
}

func TestBodyContainsRule_Match(t *testing.T) {
	rule := &BodyContainsRule{Value: "models"}

	body := []byte(`{"models": []}`)
	assert.True(t, rule.Match(nil, body))

	body = []byte(`{"error": "not found"}`)
	assert.False(t, rule.Match(nil, body))
}

func TestBodyContainsRule_Match_WithNot(t *testing.T) {
	rule := &BodyContainsRule{
		BaseRule: BaseRule{Not: true},
		Value:    "<!DOCTYPE html",
	}

	body := []byte(`OK`)
	assert.True(t, rule.Match(nil, body)) // NOT contains html = true

	body = []byte(`<!DOCTYPE html><body>OK</body>`)
	assert.False(t, rule.Match(nil, body)) // NOT contains html = false
}

func TestBodyPrefixRule_Match(t *testing.T) {
	rule := BodyPrefixRule{Value: "OK"}

	body := []byte(`OK - server is running`)
	assert.True(t, rule.Match(nil, body))

	body = []byte(`Error: server down`)
	assert.False(t, rule.Match(nil, body))
}

func TestBodyPrefixRule_Match_WithNot(t *testing.T) {
	rule := BodyPrefixRule{
		BaseRule: BaseRule{Not: true},
		Value:    "Error",
	}

	body := []byte(`OK`)
	assert.True(t, rule.Match(nil, body))

	body = []byte(`Error: something went wrong`)
	assert.False(t, rule.Match(nil, body))
}

func TestHeaderContainsRule_Match(t *testing.T) {
	rule := HeaderContainsRule{Header: "Server", Value: "uvicorn"}

	resp := &http.Response{Header: http.Header{"Server": []string{"uvicorn/0.18.0"}}}
	assert.True(t, rule.Match(resp, nil))

	resp = &http.Response{Header: http.Header{"Server": []string{"nginx"}}}
	assert.False(t, rule.Match(resp, nil))
}

func TestHeaderContainsRule_Match_WithNot(t *testing.T) {
	rule := HeaderContainsRule{
		BaseRule: BaseRule{Not: true},
		Header:   "Content-Type",
		Value:    "text/html",
	}

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}}
	assert.True(t, rule.Match(resp, nil))

	resp = &http.Response{Header: http.Header{"Content-Type": []string{"text/html"}}}
	assert.False(t, rule.Match(resp, nil))
}

func TestHeaderPrefixRule_Match(t *testing.T) {
	rule := HeaderPrefixRule{Header: "Server", Value: "llama"}

	resp := &http.Response{Header: http.Header{"Server": []string{"llama.cpp"}}}
	assert.True(t, rule.Match(resp, nil))

	resp = &http.Response{Header: http.Header{"Server": []string{"nginx"}}}
	assert.False(t, rule.Match(resp, nil))
}

func TestUnmarshalRule(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantType string
		wantNot  bool
	}{
		{
			name:     "status rule",
			yaml:     "type: status\nvalue: 200",
			wantType: "status",
		},
		{
			name:     "body.contains rule",
			yaml:     "type: body.contains\nvalue: models",
			wantType: "body.contains",
		},
		{
			name:     "body.contains with not",
			yaml:     "type: body.contains\nvalue: error\nnot: true",
			wantType: "body.contains",
			wantNot:  true,
		},
		{
			name:     "body.prefix rule",
			yaml:     "type: body.prefix\nvalue: OK",
			wantType: "body.prefix",
		},
		{
			name:     "header.contains rule",
			yaml:     "type: header.contains\nheader: Server\nvalue: uvicorn",
			wantType: "header.contains",
		},
		{
			name:     "header.prefix rule",
			yaml:     "type: header.prefix\nheader: Server\nvalue: llama",
			wantType: "header.prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw RawRule
			err := yaml.Unmarshal([]byte(tt.yaml), &raw)
			require.NoError(t, err)

			rule, err := raw.ToRule()
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, rule.GetType())
			assert.Equal(t, tt.wantNot, rule.IsNegated())
		})
	}
}

func TestContentTypeRule_Match(t *testing.T) {
	tests := []struct {
		name        string
		ruleValue   string
		contentType string
		not         bool
		want        bool
	}{
		{
			name:        "exact match",
			ruleValue:   "application/json",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "match with charset",
			ruleValue:   "application/json",
			contentType: "application/json; charset=utf-8",
			want:        true,
		},
		{
			name:        "case insensitive",
			ruleValue:   "application/json",
			contentType: "Application/JSON",
			want:        true,
		},
		{
			name:        "no match",
			ruleValue:   "application/json",
			contentType: "text/html",
			want:        false,
		},
		{
			name:        "negated - not html",
			ruleValue:   "text/html",
			contentType: "application/json",
			not:         true,
			want:        true,
		},
		{
			name:        "negated - is html",
			ruleValue:   "text/html",
			contentType: "text/html; charset=utf-8",
			not:         true,
			want:        false,
		},
		{
			name:        "empty content-type header",
			ruleValue:   "application/json",
			contentType: "",
			want:        false,
		},
		{
			name:        "nil response",
			ruleValue:   "application/json",
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &ContentTypeRule{
				BaseRule: BaseRule{Type: "content-type", Not: tt.not},
				Value:    tt.ruleValue,
			}

			var resp *http.Response
			if tt.name != "nil response" {
				resp = &http.Response{Header: http.Header{}}
				if tt.contentType != "" {
					resp.Header.Set("Content-Type", tt.contentType)
				}
			}

			assert.Equal(t, tt.want, rule.Match(resp, nil))
		})
	}
}

func TestContentTypeRule_Unmarshal(t *testing.T) {
	yamlStr := "type: content-type\nvalue: application/json"
	
	var raw RawRule
	err := yaml.Unmarshal([]byte(yamlStr), &raw)
	require.NoError(t, err)

	rule, err := raw.ToRule()
	require.NoError(t, err)
	assert.Equal(t, "content-type", rule.GetType())
	
	contentTypeRule, ok := rule.(*ContentTypeRule)
	require.True(t, ok)
	assert.Equal(t, "application/json", contentTypeRule.Value)
}
