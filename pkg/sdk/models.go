package sdk

import (
	"encoding/json"

	"github.com/praetorian-inc/julius/pkg/types"
)

const (
	categoryMCP    = "mcp"
	webpageTypeLLM = "llm"
	webpageTypeMCP = "mcp"
)

type GeneratorConfig struct {
	Source       string            `json:"source,omitempty"`
	Type         string            `json:"type"`
	Endpoint     string            `json:"endpoint"`
	APIKey       string            `json:"api_key,omitempty"`
	Model        string            `json:"model,omitempty"`
	Method       string            `json:"method,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         string            `json:"body,omitempty"`
	ResponsePath string            `json:"response_path,omitempty"`
	ResponseType string            `json:"content_type,omitempty"`
	Proxy        string            `json:"proxy,omitempty"`
	Timeout      int               `json:"timeout,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

type Webpage struct {
	URL              string            `json:"url"`
	Type             string            `json:"type,omitempty"`
	Service          string            `json:"service,omitempty"`
	GeneratorConfigs []GeneratorConfig `json:"generator_configs,omitempty"`
	Specificity      int               `json:"specificity,omitempty"`
	Parent           WebApplication    `json:"parent"`
}

type WebApplication struct {
	PrimaryURL string `json:"primary_url"`
	Name       string `json:"name"`
	Status     string `json:"status,omitempty"`
}

func webpageType(category string) string {
	if category == categoryMCP {
		return webpageTypeMCP
	}
	return webpageTypeLLM
}

func toGeneratorConfigs(configs []types.GeneratorConfig) []GeneratorConfig {
	if len(configs) == 0 {
		return nil
	}
	data, err := json.Marshal(configs)
	if err != nil {
		return nil
	}
	var out []GeneratorConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	for i := range out {
		out[i].Source = "julius"
	}
	return out
}
