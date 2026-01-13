package scanner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/itchyny/gojq"
)

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
