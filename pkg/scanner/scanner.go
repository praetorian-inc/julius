package scanner

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
)

type Scanner struct {
	client *http.Client
	cache  map[string]*CachedResponse
}

func NewScanner(timeout time.Duration) *Scanner {
	return &Scanner{
		client: &http.Client{
			Timeout: timeout,
		},
		cache: make(map[string]*CachedResponse),
	}
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

func (s *Scanner) Scan(target string, probes []*types.ProbeDefinition) *types.Result {
	for _, pd := range probes {
		for _, p := range pd.Probes {
			p.ApplyDefaults()

			matched, err := s.Probe(target, p)
			if err != nil {
				continue
			}

			if !matched {
				continue
			}

			result := &types.Result{
				Target:       target,
				Service:      pd.Name,
				Confidence:   p.Confidence,
				MatchedProbe: p.Path,
				Category:     pd.Category,
			}

			if pd.Models == nil {
				return result
			}
			models, err := s.fetchModels(target, pd.Models)
			if err != nil {
				result.Error = err.Error()
			}
			result.Models = models

			return result
		}
	}

	return nil
}

func (s *Scanner) Probe(target string, p types.Probe) (bool, error) {
	resp, body, err := s.doRequest(target, p.Method, p.Path, p.Body, p.Headers)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}

	rules, err := p.GetRules()
	if err != nil {
		return false, fmt.Errorf("parsing rules: %w", err)
	}

	matched := probe.MatchRules(resp, body, rules)
	return matched, nil
}

func (s *Scanner) fetchModels(target string, cfg *types.ModelsConfig) ([]string, error) {
	resp, body, err := s.doRequest(target, cfg.Method, cfg.Path, cfg.Body, cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("models request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models request returned %d", resp.StatusCode)
	}

	return extractModels(body, cfg.Extract)
}

func (s *Scanner) doRequest(target, method, path, body string, headers map[string]string) (*http.Response, []byte, error) {
	if method == "" {
		method = "GET"
	}

	url := target + path

	var bodyBytes []byte
	var bodyReader io.Reader
	if body != "" {
		bodyBytes = []byte(body)
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return s.cachedRequest(req, bodyBytes)
}
