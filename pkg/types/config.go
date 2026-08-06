package types

type AugustusConfig struct {
	Generator      string          `yaml:"generator"`
	ConfigTemplate GeneratorConfig `yaml:"config_template"`
}

type GeneratorConfig struct {
	Type         string            `yaml:"type" json:"type"`
	Endpoint     string            `yaml:"endpoint" json:"endpoint"`
	APIKey       string            `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	Model        string            `yaml:"model,omitempty" json:"model,omitempty"`
	Method       string            `yaml:"method,omitempty" json:"method,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body         string            `yaml:"body,omitempty" json:"body,omitempty"`
	ResponsePath string            `yaml:"response_path,omitempty" json:"response_path,omitempty"`
	ResponseType string            `yaml:"content_type,omitempty" json:"content_type,omitempty"`
	Proxy        string            `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	Timeout      int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	// Extra is a generator-type-specific passthrough for config keys that have no
	// dedicated field above (which are shaped for HTTP/REST LLM generators). It
	// exists for generators like the augustus MCP generator, whose config is
	// carried entirely in extra keys (transport, mode). Values support the same
	// $TARGET/$MODEL substitution as the typed fields — with the same caveat that
	// $MODEL is only substituted when the probe declares `models:`; with no models
	// (e.g. mcp-server) a `$MODEL`-valued extra stays the literal `$MODEL`, so
	// don't rely on it there. Like Headers, extra values are NOT rewritten by
	// --base-paths (only Endpoint is), so avoid URL-valued extras.
	// Values are strings only: a non-string YAML value (e.g. `verify_tls: false`)
	// is coerced to its string form, which a bool/int-expecting generator reads as
	// its default with no error. map[string]any is the eventual home for typed
	// keys (LAB-4935).
	Extra map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
}
