package scanner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
)

var ansiEscRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func sanitizeMetadataValue(s string) string {
	s = ansiEscRe.ReplaceAllString(s, "")

	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
	}
	s = b.String()

	if len(s) > 256 {
		s = s[:256]
	}
	return s
}

func NormalizeTarget(target string) string {
	target = strings.TrimSpace(target)

	if target == "" {
		return ""
	}

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return strings.TrimRight(target, "/")
	}

	// Strip trailing slashes from bare hosts (path is only slashes or empty).
	// Preserve trailing slashes on real paths so empty-path probes
	// test the exact URL supplied (e.g. https://host/mcp/).
	if strings.Trim(u.Path, "/") == "" {
		u.Path = ""
		return u.String()
	}

	return target
}

func NormalizeTargets(targets []string) []string {
	var normalized []string
	for _, t := range targets {
		n := NormalizeTarget(t)
		if n != "" {
			normalized = append(normalized, n)
		}
	}
	return normalized
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

// stripSSEFraming extracts the JSON payload from SSE-framed responses.
// MCP servers may respond with text/event-stream wrapping a JSON-RPC body
// in a "data:" line. If no SSE framing is detected, returns body unchanged.
func stripSSEFraming(body []byte) []byte {
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "data:") {
			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if len(payload) > 0 && payload[0] == '{' {
				return []byte(payload)
			}
		}
	}
	return body
}

func extractFromBody(body []byte, extracts map[string]string) map[string]string {
	body = stripSSEFraming(body)

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}

	result := make(map[string]string, len(extracts))
	for field, jqExpr := range extracts {
		query, err := gojq.Parse(jqExpr)
		if err != nil {
			continue
		}
		iter := query.Run(data)
		v, ok := iter.Next()
		if !ok {
			continue
		}
		if _, isErr := v.(error); isErr {
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			result[field] = sanitizeMetadataValue(s)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
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
