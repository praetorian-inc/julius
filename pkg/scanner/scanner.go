package scanner

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

	req, err := http.NewRequest(p.Method, url, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	matched := probe.Match(resp, p.Match)
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
