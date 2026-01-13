// pkg/types/rules_test.go
package types

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRule_Match(t *testing.T) {
	rule := StatusRule{
		BaseRule: BaseRule{
			Type: "status",
			Not:  false,
		},
		Status: 200,
	}

	// Create mock response with status 200
	resp := &http.Response{
		StatusCode: 200,
	}

	// Should match when status is 200
	assert.True(t, rule.Match(resp, nil), "StatusRule should match when status code is 200")

	// Should not match when status is different
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

	// Create mock response with status 200
	resp := &http.Response{
		StatusCode: 200,
	}

	// Should match when status is NOT 404 (negated)
	assert.True(t, rule.Match(resp, nil), "StatusRule with Not=true should match when status is NOT 404")

	// Should not match when status IS 404 (negated)
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
