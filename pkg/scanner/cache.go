package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
)

type CachedResponse struct {
	Response *http.Response
	Body     []byte
	Err      error
}

func cacheKey(method, url string, headers http.Header, body []byte) string {
	h := md5.New()
	h.Write([]byte(method))
	h.Write([]byte(url))

	// Sort header keys for determinism
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(strings.Join(headers[k], ",")))
	}

	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Scanner) cachedRequest(req *http.Request, body []byte) (*http.Response, []byte, error) {
	key := cacheKey(req.Method, req.URL.String(), req.Header, body)

	// Use singleflight to deduplicate concurrent requests
	result, err, _ := s.inflight.Do(key, func() (any, error) {
		if cached, ok := s.cache.Load(key); ok {
			return cached, nil
		}

		reqCopy := req.Clone(req.Context())
		if body != nil {
			reqCopy.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		resp, err := s.client.Do(reqCopy)
		if err != nil {
			slog.Error("Getting response", "method", req.Method, "url", req.URL.String(), "err", err)
			cached := &CachedResponse{Err: err}
			s.cache.Store(key, cached)
			return cached, nil
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, s.maxResponseSize))
		resp.Body.Close()
		if err != nil {
			slog.Error("Reading response body", "method", req.Method, "url", req.URL.String(), "err", err)
			cached := &CachedResponse{Err: err}
			s.cache.Store(key, cached)
			return cached, nil
		}

		if int64(len(respBody)) == s.maxResponseSize {
			slog.Warn("Response body truncated at size limit", "method", req.Method, "url", req.URL.String(), "limit", s.maxResponseSize)
		}

		resp.Body = nil // Clear to make it obvious this shouldn't be read

		cached := &CachedResponse{Response: resp, Body: respBody}
		s.cache.Store(key, cached)

		return cached, nil
	})

	if err != nil {
		return nil, nil, err
	}

	cached := result.(*CachedResponse)
	if cached.Err != nil {
		return nil, nil, cached.Err
	}
	return cached.Response, cached.Body, nil
}
