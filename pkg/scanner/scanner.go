package scanner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
)

type Scanner struct {
	client      *http.Client
	cache       sync.Map
	inflight    singleflight.Group
	concurrency int
}

func NewScanner(timeout time.Duration, concurrency int) *Scanner {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &Scanner{
		client: &http.Client{
			Timeout: timeout,
		},
		concurrency: concurrency,
	}
}

func (s *Scanner) ScanAll(targets []string, probes []*types.Probe, augustus bool) []types.Result {
	var results []types.Result

	for _, target := range targets {
		targetResults := s.Scan(target, probes, augustus)
		results = append(results, targetResults...)
	}

	return results
}

func (s *Scanner) Scan(target string, probes []*types.Probe, augustus bool) []types.Result {
	var (
		results   []types.Result
		resultsMu sync.Mutex
	)

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(s.concurrency)

	for _, p := range probes {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			matched, matchedReq := s.matchProbe(target, p)
			if !matched {
				return nil
			}

			result := types.Result{
				Target:         target + matchedReq.Path,
				Service:        p.Name,
				MatchedRequest: matchedReq.Path,
				Category:       p.Category,
				Specificity:    p.GetSpecificity(),
			}

			if p.Models != nil {
				models, err := s.fetchModels(target, p.Models)
				if err != nil {
					result.Error = err.Error()
				}
				result.Models = models
			}

			if augustus {
				result.GeneratorConfigs = p.BuildGeneratorConfigs(target, result.Models)
			}

			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()

			return nil
		})
	}

	g.Wait()

	// Sort by specificity (highest first)
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Specificity > results[j].Specificity
	})

	return results
}

func (s *Scanner) matchProbe(target string, p *types.Probe) (bool, types.Request) {
	if p.RequiresAll() {
		return s.matchProbeAll(target, p)
	}
	return s.matchProbeAny(target, p)
}

func (s *Scanner) matchProbeAny(target string, p *types.Probe) (bool, types.Request) {
	for _, req := range p.Requests {
		req.ApplyDefaults()

		matched, err := s.DoRequest(target, req)
		if err != nil || !matched {
			continue
		}

		return true, req
	}

	return false, types.Request{}
}

func (s *Scanner) matchProbeAll(target string, p *types.Probe) (bool, types.Request) {
	if len(p.Requests) == 0 {
		return false, types.Request{}
	}

	var firstReq types.Request
	for i, req := range p.Requests {
		req.ApplyDefaults()

		matched, err := s.DoRequest(target, req)
		if err != nil || !matched {
			return false, types.Request{}
		}

		if i == 0 {
			firstReq = req
		}
	}

	return true, firstReq
}

func (s *Scanner) DoRequest(target string, req types.Request) (bool, error) {
	resp, body, err := s.doHTTPRequest(target, req.Method, req.Path, req.Body, req.Headers)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}

	rules, err := req.GetRules()
	if err != nil {
		return false, fmt.Errorf("parsing rules: %w", err)
	}

	matched := probe.MatchRules(resp, body, rules)
	return matched, nil
}

func (s *Scanner) fetchModels(target string, cfg *types.ModelsConfig) ([]string, error) {
	resp, body, err := s.doHTTPRequest(target, cfg.Method, cfg.Path, cfg.Body, cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("models request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models request returned %d", resp.StatusCode)
	}

	return extractModels(body, cfg.Extract)
}

func (s *Scanner) doHTTPRequest(target, method, path, body string, headers map[string]string) (*http.Response, []byte, error) {
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
