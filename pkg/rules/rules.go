package rules

import (
	"fmt"
	"net/http"
)

type Rule interface {
	Match(resp *http.Response, body []byte) bool
	GetType() string
	IsNegated() bool
}

type BaseRule struct {
	Type string
	Not  bool
}

func (b BaseRule) GetType() string {
	return b.Type
}

func (b BaseRule) IsNegated() bool {
	return b.Not
}

// RawRule is the YAML representation before conversion to typed Rule
type RawRule struct {
	Type   string `yaml:"type"`
	Not    bool   `yaml:"not,omitempty"`
	Value  any    `yaml:"value,omitempty"`
	Header string `yaml:"header,omitempty"`
}

type Decoder func(raw *RawRule) (Rule, error)

var ruleDecoders = map[string]Decoder{}

func Register(typeName string, decoder Decoder) {
	ruleDecoders[typeName] = decoder
}

func (r *RawRule) ToRule() (Rule, error) {
	decoder, ok := ruleDecoders[r.Type]
	if !ok {
		return nil, fmt.Errorf("unknown rule type: %s", r.Type)
	}
	return decoder(r)
}

func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case uint64:
		return int(val), nil
	default:
		return 0, fmt.Errorf("value must be int, got %T", v)
	}
}

func toString(v any) (string, error) {
	val, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value must be string, got %T", v)
	}
	return val, nil
}
