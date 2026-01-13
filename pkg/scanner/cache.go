package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
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

	if cached, ok := s.cache[key]; ok {
		return cached.Response, cached.Body, cached.Err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Printf("[cache] %s %s: %v\n", req.Method, req.URL.Path, err)
		s.cache[key] = &CachedResponse{Err: err}
		return nil, nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Printf("[cache] %s %s: failed to read body: %v\n", req.Method, req.URL.Path, err)
		s.cache[key] = &CachedResponse{Err: err}
		return nil, nil, err
	}

	resp.Body = nil // Clear to make it obvious this shouldn't be read

	s.cache[key] = &CachedResponse{Response: resp, Body: respBody}
	return resp, respBody, nil
}
