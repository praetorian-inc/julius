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
// Before the fix, deepspeed-mii's vector 2 (legacy Flask gateway, POST
// /mii/deepspeed-mii) matched any response body containing the bare
// substring `"text"`, with no status or content-type gate. Any JSON
// endpoint that happened to return a `text` field would be misidentified
// as deepspeed-mii. The fix adds `status: 200` and
// `content-type: application/json` gates ahead of the `body.contains`
// rule, mirroring vector 1's idiom.
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
	require.Len(t, deepspeedMII.Requests, 2, "deepspeed-mii should have 2 request vectors")

	vector2 := deepspeedMII.Requests[1]
	require.Equal(t, "/mii/deepspeed-mii", vector2.Path, "vector 2 must be the legacy Flask gateway path")

	ruleList, err := vector2.GetRules()
	require.NoError(t, err, "vector 2 rules should parse without error")

	tests := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		wantMatch   bool
	}{
		{
			name:        "200 application/json with text field matches",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"text": "generated output"}`,
			wantMatch:   true,
		},
		{
			name:        "200 application/json with charset matches",
			statusCode:  200,
			contentType: "application/json; charset=utf-8",
			body:        `{"text": "generated output"}`,
			wantMatch:   true,
		},
		{
			name:        "200 text/plain with text field does not match (pre-fix regression case)",
			statusCode:  200,
			contentType: "text/plain",
			body:        `{"text": "generated output"}`,
			wantMatch:   false,
		},
		{
			name:        "500 application/json with text field does not match",
			statusCode:  500,
			contentType: "application/json",
			body:        `{"text": "generated output"}`,
			wantMatch:   false,
		},
		{
			name:        "200 application/json without text field does not match",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"result": "generated output"}`,
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
