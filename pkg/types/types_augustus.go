//go:build augustus

package types

import (
	"strings"

	"github.com/praetorian-inc/augustus/pkg/generator"
)

// GeneratorConfig is the Augustus generator configuration
type GeneratorConfig = generator.Config

// AugustusConfig defines how to scan this service with Augustus.
// ConfigTemplate uses $TARGET and $MODEL placeholders that get resolved at runtime.
type AugustusConfig struct {
	Generator      string          `yaml:"generator"`
	ConfigTemplate GeneratorConfig `yaml:"config_template"`
}

// BuildGeneratorConfigs builds Augustus configs from probe definition
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
