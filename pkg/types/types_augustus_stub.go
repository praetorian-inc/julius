//go:build !augustus

package types

// GeneratorConfig is a stub type when Augustus is not enabled
type GeneratorConfig map[string]any

// AugustusConfig is ignored when Augustus is not enabled
type AugustusConfig struct {
	Generator      string          `yaml:"generator"`
	ConfigTemplate GeneratorConfig `yaml:"config_template"`
}

// BuildGeneratorConfigs returns nil when Augustus is not enabled
func (p *Probe) BuildGeneratorConfigs(target string, models []string) []GeneratorConfig {
	return nil
}
