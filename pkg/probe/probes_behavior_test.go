package probe

import (
	"net/http"
	"testing"

	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/praetorian-inc/julius/probes"
	"github.com/stretchr/testify/require"
)

// TestDeepSpeedMII_Vector2RequiresJSONGate is a regression test for LAB-4680.
//
// Before the fix, deepspeed-mii's vector 2 (Flask gateway, POST
// /mii/deepspeed-mii) matched any response body containing the bare
// substring `"text"`. That marker was wrong twice over: it was generic
// enough to match unrelated JSON services (e.g. an MCP response shaped
// like `{"content": [{"type": "text", "text": "..."}]}`), causing the
// false positive this ticket fixes, AND it never matched a real
// DeepSpeed-MII instance in the first place. The gateway actually returns
// a JSON array of Response.to_msg_dict() (mii/batching/data_classes.py)
// with fields `generated_text`, `prompt_length`, `generated_length`, and
// `finish_reason` - note `"text"` is NOT a substring of `"generated_text"`
// (the character before `text` is `_`, not a quote). The only upstream
// `text` field lives on CompletionResponseChoice, which serves
// /v1/completions, a path this vector does not probe.
//
// The fix re-anchors vector 2 on `status: 200`, `content-type:
// application/json`, and all four of the real response fields ANDed
// together via body.contains rules, mirroring vector 1's idiom.
//
// This test loads the REAL embedded probe (not a hand-built copy) so it
// pins the shipped YAML in probes/deepspeed-mii.yaml, then drives vector
// 2's actual rule set against synthetic HTTP responses.
func TestDeepSpeedMII_Vector2RequiresJSONGate(t *testing.T) {
	loadedProbes, err := LoadProbesFromFS(probes.EmbeddedProbes, ".")
	require.NoError(t, err, "LoadProbesFromFS() should not error")

	var deepspeedMII *types.Probe
	for _, p := range loadedProbes {
		if p.Name == "deepspeed-mii" {
			deepspeedMII = p
			break
		}
	}
	require.NotNil(t, deepspeedMII, "deepspeed-mii probe must be present in embedded probes")

	var vector2 types.Request
	for _, r := range deepspeedMII.Requests {
		if r.Path == "/mii/deepspeed-mii" {
			vector2 = r
			break
		}
	}
	require.NotEmpty(t, vector2.Path, "vector 2 (Flask gateway path) must be present in the probe")

	ruleList, err := vector2.GetRules()
	require.NoError(t, err, "vector 2 rules should parse without error")

	const fullMIIResponse = `[{"generated_text": "hello world", "prompt_length": 4, "generated_length": 2, "finish_reason": "length"}]`

	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		wantMatch   bool
	}{
		{
			name:        "200 application/json with full MII response matches",
			statusCode:  200,
			contentType: "application/json",
			body:        fullMIIResponse,
			wantMatch:   true,
		},
		{
			name:        "200 application/json with charset and full MII response matches",
			statusCode:  200,
			contentType: "application/json; charset=utf-8",
			body:        fullMIIResponse,
			wantMatch:   true,
		},
		{
			name:        "200 text/plain with full MII response does not match (content-type gate)",
			statusCode:  200,
			contentType: "text/plain",
			body:        fullMIIResponse,
			wantMatch:   false,
		},
		{
			name:        "500 application/json with full MII response does not match (status gate)",
			statusCode:  500,
			contentType: "application/json",
			body:        fullMIIResponse,
			wantMatch:   false,
		},
		{
			name:        "200 application/json with MCP-shaped bare text field does not match (LAB-4680 false-positive regression)",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"content": [{"type": "text", "text": "hello"}]}`,
			wantMatch:   false,
		},
		{
			name:        "200 application/json with only generated_text field does not match (fields ANDed)",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"generated_text": "hello world"}`,
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     http.Header{"Content-Type": []string{tt.contentType}},
			}
			body := []byte(tt.body)

			matched := MatchRules(resp, body, ruleList)
			require.Equal(t, tt.wantMatch, matched)
		})
	}
}
