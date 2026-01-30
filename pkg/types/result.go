package types

type Result struct {
	Target           string            `json:"target"`
	Service          string            `json:"service"`
	MatchedRequest   string            `json:"matched_request"`
	Category         string            `json:"category"`
	Specificity      int               `json:"specificity"`
	Models           []string          `json:"models,omitempty"`
	GeneratorConfigs []GeneratorConfig `json:"generator_configs,omitempty"`
	Error            string            `json:"error,omitempty"`
}

type OutputWriter interface {
	Write(results []Result) error
}
