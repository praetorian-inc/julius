package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
)

type Scanner struct {
	client *http.Client
}

func NewScanner(timeout time.Duration) *Scanner {
	return &Scanner{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *Scanner) Probe(target string, p types.Probe) (bool, error) {
	url := target + p.Path

	// Create request body if specified
	var bodyReader io.Reader
	if p.Body != "" {
		bodyReader = strings.NewReader(p.Body)
	}

	req, err := http.NewRequest(p.Method, url, bodyReader)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	// Set custom headers
	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading response body: %w", err)
	}

	rules, err := p.GetRules()
	if err != nil {
		return false, fmt.Errorf("parsing rules: %w", err)
	}

	matched := probe.MatchRules(resp, body, rules)
	return matched, nil
}

func (s *Scanner) Scan(target string, probes []*types.ProbeDefinition) *types.Result {
	for _, pd := range probes {
		for _, p := range pd.Probes {
			p.ApplyDefaults()

			matched, err := s.Probe(target, p)
			if err != nil {
				continue
			}

			if matched {
				return &types.Result{
					Target:       target,
					Service:      pd.Name,
					Confidence:   p.Confidence,
					MatchedProbe: p.Path,
					Category:     pd.Category,
				}
			}
		}
	}

	return nil
}

func (s *Scanner) ScanAll(targets []string, probes []*types.ProbeDefinition) []types.Result {
	var results []types.Result

	for _, target := range targets {
		result := s.Scan(target, probes)
		if result != nil {
			results = append(results, *result)
		}
	}

	return results
}

func ExtractPort(target string) int {
	u, err := url.Parse(target)
	if err != nil {
		return 0
	}

	port := u.Port()
	if port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return 0
		}
		return p
	}

	switch u.Scheme {
	case "https":
		return 443
	case "http":
		return 80
	default:
		return 0
	}
}
