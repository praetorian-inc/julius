// pkg/scanner/fixture_test.go
//
// Golden-file HTTP fixture harness for Julius probes.
//
// Each probe in probes/*.yaml has a companion fixture in
// pkg/scanner/testdata/fixtures/<probe>.yaml describing one or more cases. A
// case defines canned HTTP responses keyed by request path plus the expected
// overall match result (true/false). For each case the harness serves those
// responses from an httptest.Server and runs the *real* probe (loaded from the
// embedded probe library) through the *real* scanner, then asserts that the
// probe matched iff the case said it should.
//
// Because the shipped probe is exercised end-to-end, this catches match-rule
// regressions that `julius validate` (schema-only) cannot: a changed status
// code, a renamed body marker, a flipped negation, or a broken require:all/any
// chain will fail the corresponding fixture case.
//
// To add coverage for a new probe, drop a <probe>.yaml fixture in the fixtures
// directory with at least one positive and one negative case. The completeness
// gate (TestEveryProbeHasFixture) fails the build if any shipped probe has no
// fixture.
package scanner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/praetorian-inc/julius/probes"
)

const fixturesDir = "testdata/fixtures"

// fixture is the golden file for a single probe: probes/<probe>.yaml is
// validated against the cases here.
type fixture struct {
	Probe string        `yaml:"probe"`
	Cases []fixtureCase `yaml:"cases"`
	path  string        // source file, for error messages
}

// fixtureCase is one scenario: a set of canned responses and the match result
// the probe must produce against them.
type fixtureCase struct {
	Name      string                     `yaml:"name"`
	Match     bool                       `yaml:"match"`
	Responses map[string]fixtureResponse `yaml:"responses"`
}

// fixtureResponse is the canned HTTP response served for a given request path.
// Method, when set, restricts the response to requests using that HTTP method;
// a mismatch is served 405 so a probe whose method regresses (e.g. POST->GET)
// no longer receives the canned body. Empty means any method is accepted.
type fixtureResponse struct {
	Status  int               `yaml:"status"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

// handler serves the case's canned responses keyed by the full request URI
// (path plus ?query), falling back to the bare path if no exact match is
// found. Any path not present in the fixture returns 404 with an empty body,
// which lets negative cases (and unmet require:all chains) fail naturally.
func (c fixtureCase) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI() // path plus ?query, matching how probes specify paths
		resp, ok := c.Responses[key]
		if !ok {
			resp, ok = c.Responses[r.URL.Path]
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if resp.Method != "" && !strings.EqualFold(r.Method, resp.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		status := resp.Status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, resp.Body)
	})
}

// loadProbeMap loads every embedded probe, keyed by name. This is exactly the
// probe set that ships in the binary.
func loadProbeMap(t *testing.T) map[string]*types.Probe {
	t.Helper()
	loaded, err := probe.LoadProbesFromFS(probes.EmbeddedProbes, ".")
	require.NoError(t, err, "loading embedded probes")
	require.NotEmpty(t, loaded, "embedded probe set should not be empty")

	m := make(map[string]*types.Probe, len(loaded))
	for _, p := range loaded {
		require.NotContains(t, m, p.Name, "duplicate probe name %q", p.Name)
		m[p.Name] = p
	}
	return m
}

// loadFixtures reads and parses every fixture YAML in the fixtures directory.
func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err, "reading fixtures directory")

	var fixtures []fixture
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(fixturesDir, e.Name())
		data, err := os.ReadFile(path)
		require.NoError(t, err, "reading %s", path)

		var fx fixture
		require.NoError(t, yaml.Unmarshal(data, &fx), "parsing %s", path)
		fx.path = path
		fixtures = append(fixtures, fx)
	}
	require.NotEmpty(t, fixtures, "no fixtures found in %s", fixturesDir)
	return fixtures
}

// TestProbeFixtures runs every fixture case against its real probe and asserts
// the match result. This is the core behavioral gate.
func TestProbeFixtures(t *testing.T) {
	t.Parallel()
	probeMap := loadProbeMap(t)

	for _, fx := range loadFixtures(t) {
		p, ok := probeMap[fx.Probe]
		if !assert.Truef(t, ok, "%s references unknown probe %q", fx.path, fx.Probe) {
			continue
		}
		if !assert.NotEmptyf(t, fx.Cases, "%s has no cases", fx.path) {
			continue
		}

		t.Run(fx.Probe, func(t *testing.T) {
			t.Parallel()
			for _, c := range fx.Cases {
				if !assert.NotEmptyf(t, c.Name, "%s has a case with no name", fx.path) {
					continue
				}
				if !assert.NotEmptyf(t, c.Responses, "%s/%s has no responses", fx.Probe, c.Name) {
					continue
				}

				t.Run(c.Name, func(t *testing.T) {
					t.Parallel()
					server := httptest.NewServer(c.handler())
					defer server.Close()

					s := NewScanner(WithTimeout(5 * time.Second))
					results := s.Scan(server.URL, []*types.Probe{p}, false)
					matched := len(results) > 0

					assert.Equalf(t, c.Match, matched,
						"probe %q case %q: expected match=%v, got match=%v",
						fx.Probe, c.Name, c.Match, matched)
				})
			}
		})
	}
}

// TestEveryProbeHasFixture is the completeness gate: every shipped probe must
// have a fixture with at least one positive and one negative case. Adding a
// probe without a fixture fails the build.
func TestEveryProbeHasFixture(t *testing.T) {
	t.Parallel()
	probeMap := loadProbeMap(t)

	type coverage struct{ hasPositive, hasNegative bool }
	covered := make(map[string]coverage)
	for _, fx := range loadFixtures(t) {
		c := covered[fx.Probe]
		for _, fc := range fx.Cases {
			if fc.Match {
				c.hasPositive = true
			} else {
				c.hasNegative = true
			}
		}
		covered[fx.Probe] = c
	}

	for name := range probeMap {
		c, ok := covered[name]
		if !assert.Truef(t, ok, "probe %q has no fixture in %s", name, fixturesDir) {
			continue
		}
		assert.Truef(t, c.hasPositive, "probe %q fixture has no positive (match: true) case", name)
		assert.Truef(t, c.hasNegative, "probe %q fixture has no negative (match: false) case", name)
	}
}
