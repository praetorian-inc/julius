package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripSSEFraming(t *testing.T) {
	const initJSON = `{"jsonrpc":"2.0","result":{"protocolVersion":"2025-06-18"}}`

	tests := []struct {
		name string
		body string
		want string
	}{
		{"plain JSON unchanged", initJSON, initJSON},
		{"SSE-wrapped JSON", "event: message\ndata: " + initJSON + "\n\n", initJSON},
		{"first JSON data line wins", "data: {\"a\":1}\ndata: {\"b\":2}\n\n", `{"a":1}`},
		{"empty body", "", ""},
		{"non-JSON data line unchanged", "data: hello", "data: hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(stripSSEFraming([]byte(tt.body))))
		})
	}
}

func TestExtractFromBody(t *testing.T) {
	const versionExpr = `.result.protocolVersion // .result.supportedVersions[0]`
	versionExtract := map[string]string{"protocol_version": versionExpr}

	const initJSON = `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","serverInfo":{"name":"test","version":"1.0"}}}`

	tests := []struct {
		name     string
		body     string
		extracts map[string]string
		want     map[string]string
	}{
		{
			name:     "discover response (V1)",
			body:     `{"jsonrpc":"2.0","id":1,"result":{"supportedVersions":["2026-07-28"],"_meta":{"io.modelcontextprotocol/serverInfo":{"name":"test","version":"1.0"}}}}`,
			extracts: versionExtract,
			want:     map[string]string{"protocol_version": "2026-07-28"},
		},
		{
			name:     "initialize response (V3/V1b)",
			body:     initJSON,
			extracts: versionExtract,
			want:     map[string]string{"protocol_version": "2025-06-18"},
		},
		{
			name:     "SSE-wrapped response",
			body:     "event: message\ndata: " + initJSON + "\n\n",
			extracts: versionExtract,
			want:     map[string]string{"protocol_version": "2025-06-18"},
		},
		{
			name:     "empty body (OAuth 401)",
			body:     "",
			extracts: versionExtract,
			want:     nil,
		},
		{
			name:     "invalid JSON",
			body:     "not json",
			extracts: versionExtract,
			want:     nil,
		},
		{
			name: "multiple extractions",
			body: initJSON,
			extracts: map[string]string{
				"protocol_version": ".result.protocolVersion",
				"server_name":      ".result.serverInfo.name",
			},
			want: map[string]string{"protocol_version": "2025-06-18", "server_name": "test"},
		},
		{
			name: "partial match",
			body: initJSON,
			extracts: map[string]string{
				"protocol_version": ".result.protocolVersion",
				"missing":          ".result.nope",
			},
			want: map[string]string{"protocol_version": "2025-06-18"},
		},
		{
			name:     "non-string result omitted",
			body:     `{"result":{"count":3}}`,
			extracts: map[string]string{"count": ".result.count"},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractFromBody([]byte(tt.body), tt.extracts))
		})
	}
}
