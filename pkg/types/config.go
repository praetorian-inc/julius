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
}
