package types

import "strings"

const (
	SpecificityGeneric = 1   // Fallback probes (lowest priority)
	SpecificityLow     = 25  // Broad detection
	SpecificityMedium  = 50  // Default
	SpecificityHigh    = 75  // Service-specific markers
	SpecificityExact   = 100 // Definitive identification
)

const (
	RequireAny = "any" // Default: match if ANY request succeeds
	RequireAll = "all" // Match only if ALL requests succeed
)

type Probe struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Category    string          `yaml:"category"`
	PortHint    int             `yaml:"port_hint"`
	Specificity int             `yaml:"specificity"`       // 1-100, 0 treated as default (50)
	Require     string          `yaml:"require,omitempty"` // "any" (default) or "all"
	APIDocs     string          `yaml:"api_docs"`
	Requests    []Request       `yaml:"requests"`
	Models      *ModelsConfig   `yaml:"models,omitempty"`
	Augustus    *AugustusConfig `yaml:"augustus,omitempty"`
}

func (p *Probe) RequiresAll() bool {
	return strings.ToLower(p.Require) == RequireAll
}

func (p *Probe) GetSpecificity() int {
	if p.Specificity <= 0 {
		return SpecificityMedium
	}
	return p.Specificity
}

func (p *Probe) BuildGeneratorConfigs(target string, models []string) []GeneratorConfig {
	if p.Augustus == nil {
		return nil
	}

	if len(models) == 0 {
		config := resolveGeneratorConfig(p.Augustus.ConfigTemplate, p.Augustus.Generator, target, "")
		return []GeneratorConfig{config}
	}

	configs := make([]GeneratorConfig, 0, len(models))
	for _, model := range models {
		config := resolveGeneratorConfig(p.Augustus.ConfigTemplate, p.Augustus.Generator, target, model)
		configs = append(configs, config)
	}
	return configs
}

func resolveGeneratorConfig(cfg GeneratorConfig, genType, target, model string) GeneratorConfig {
	cfg.Type = genType
	cfg.Endpoint = resolveVars(cfg.Endpoint, target, model)
	cfg.APIKey = resolveVars(cfg.APIKey, target, model)
	cfg.Model = resolveVars(cfg.Model, target, model)
	cfg.Body = resolveVars(cfg.Body, target, model)
	cfg.ResponsePath = resolveVars(cfg.ResponsePath, target, model)
	cfg.Proxy = resolveVars(cfg.Proxy, target, model)

	if cfg.Headers != nil {
		resolved := make(map[string]string, len(cfg.Headers))
		for k, v := range cfg.Headers {
			resolved[k] = resolveVars(v, target, model)
		}
		cfg.Headers = resolved
	}

	return cfg
}

func resolveVars(s, target, model string) string {
	s = strings.ReplaceAll(s, "$TARGET", target)
	if model != "" {
		s = strings.ReplaceAll(s, "$MODEL", model)
	}
	return s
}

type ModelsConfig struct {
	Path    string            `yaml:"path"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Extract string            `yaml:"extract"`
}
