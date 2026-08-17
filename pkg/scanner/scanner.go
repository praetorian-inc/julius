package scanner

import (
	"context"
	"crypto/tls"
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

const (
	DefaultConcurrency           = 10
	DefaultMaxResponseSize int64 = 10 * 1024 * 1024
)

type Scanner struct {
	client          *http.Client
	cache           sync.Map
	inflight        singleflight.Group
	concurrency     int
	maxResponseSize int64
	headers         map[string]string
}

type Option func(*Scanner)

func NewScanner(opts ...Option) *Scanner {
	s := &Scanner{
		client:          &http.Client{},
		concurrency:     DefaultConcurrency,
		maxResponseSize: DefaultMaxResponseSize,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

			matched, mr := s.matchProbe(target, p)
			if !matched {
				return nil
			}

			result := types.Result{
				Target:         target + mr.Request.Path,
				Service:        p.Name,
				MatchedRequest: mr.Request.Path,
				Category:       p.Category,
				Specificity:    p.GetSpecificity(),
				AuthRequired:   mr.StatusCode == http.StatusUnauthorized || mr.StatusCode == http.StatusForbidden,
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

	_ = g.Wait()

	// Sort by specificity (highest first)
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Specificity > results[j].Specificity
	})

	return results
}

type matchResult struct {
	Request    types.Request
	StatusCode int
}

func (s *Scanner) matchProbe(target string, p *types.Probe) (bool, matchResult) {
	if p.RequiresAll() {
		return s.matchProbeAll(target, p)
	}
	return s.matchProbeAny(target, p)
}

func (s *Scanner) matchProbeAny(target string, p *types.Probe) (bool, matchResult) {
	for _, req := range p.Requests {
		req.ApplyDefaults()

		statusCode, err := s.doRequestWithStatus(target, req)
		if err != nil || statusCode == -1 {
			continue
		}

		return true, matchResult{Request: req, StatusCode: statusCode}
	}

	return false, matchResult{}
}

func (s *Scanner) matchProbeAll(target string, p *types.Probe) (bool, matchResult) {
	if len(p.Requests) == 0 {
		return false, matchResult{}
	}

	var first matchResult
	for i, req := range p.Requests {
		req.ApplyDefaults()

		statusCode, err := s.doRequestWithStatus(target, req)
		if err != nil || statusCode == -1 {
			return false, matchResult{}
		}

		if i == 0 {
			first = matchResult{Request: req, StatusCode: statusCode}
		}
	}

	return true, first
}

func (s *Scanner) DoRequest(target string, req types.Request) (bool, error) {
	matched, _ := s.doRequestWithStatus(target, req)
	if matched == -1 {
		return false, nil
	}
	return matched > 0, nil
}

// doRequestWithStatus returns (statusCode, error) where statusCode == -1 means
// the request failed or rules didn't parse. Callers use the status code to
// derive auth signals (401/403 → auth required).
func (s *Scanner) doRequestWithStatus(target string, req types.Request) (int, error) {
	resp, body, err := s.doHTTPRequest(target, req.Method, req.Path, req.Body, req.Headers)
	if err != nil {
		return -1, fmt.Errorf("executing request: %w", err)
	}

	rules, err := req.GetRules()
	if err != nil {
		return -1, fmt.Errorf("parsing rules: %w", err)
	}

	if probe.MatchRules(resp, body, rules) {
		return resp.StatusCode, nil
	}
	return -1, nil
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

	// Apply global headers first (user-supplied via -H flag)
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}
	// Apply per-probe headers (can override global headers)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return s.cachedRequest(req, bodyBytes)
}

func WithTimeout(d time.Duration) Option {
	return func(s *Scanner) {
		s.client.Timeout = d
	}
}

func WithConcurrency(n int) Option {
	return func(s *Scanner) {
		if n > 0 {
			s.concurrency = n
		}
	}
}

func WithMaxResponseSize(n int64) Option {
	return func(s *Scanner) {
		if n > 0 {
			s.maxResponseSize = n
		}
	}
}

func WithTLSConfig(cfg *tls.Config) Option {
	return func(s *Scanner) {
		if cfg != nil {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.TLSClientConfig = cfg
			s.client.Transport = transport
		}
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(s *Scanner) {
		s.headers = headers
	}
}
