package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/itchyny/gojq"
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

func (s *Scanner) Probe(target string, p types.Probe) (bool, error) {
	url := target + p.Path

	var bodyBytes []byte
	var bodyReader io.Reader
	if p.Body != "" {
		bodyBytes = []byte(p.Body)
		bodyReader = strings.NewReader(p.Body)
	}

	req, err := http.NewRequest(p.Method, url, bodyReader)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}

	resp, body, err := s.cachedRequest(req, bodyBytes)
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

			if pd.Models != nil {
				models, err := s.fetchModels(target, pd.Models)
				if err != nil {
					result.Error = err.Error()
				}
				result.Models = models
			}

			return result
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

func (s *Scanner) fetchModels(target string, cfg *types.ModelsConfig) ([]string, error) {
	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	url := target + cfg.Path

	var bodyBytes []byte
	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyBytes = []byte(cfg.Body)
		bodyReader = strings.NewReader(cfg.Body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating models request: %w", err)
	}

	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	resp, body, err := s.cachedRequest(req, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("models request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models request returned %d", resp.StatusCode)
	}

	return extractModels(body, cfg.Extract)
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

func extractModels(body []byte, jqExpr string) ([]string, error) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression: %w", err)
	}

	var models []string
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return nil, fmt.Errorf("jq execution error: %w", err)
		}
		if s, ok := v.(string); ok {
			models = append(models, s)
		}
	}

	if models == nil {
		models = []string{}
	}

	return models, nil
}
