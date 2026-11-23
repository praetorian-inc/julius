package sdk

import (
	"fmt"
	"strings"
	"time"

	"github.com/praetorian-inc/capability-sdk/pkg/capability"
	"github.com/praetorian-inc/capability-sdk/pkg/capmodel"
	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/scanner"
	"github.com/praetorian-inc/julius/probes"
)

var _ capability.Capability[capmodel.Port] = (*Capability)(nil)

type Capability struct{}

func NewCapability() *Capability { return &Capability{} }

func (c *Capability) Name() string        { return "julius" }
func (c *Capability) Description() string { return "detects LLM services (Ollama, OpenAI, etc.) on HTTP endpoints" }
func (c *Capability) Input() any          { return capmodel.Port{} }

func (c *Capability) Parameters() []capability.Parameter {
	return []capability.Parameter{
		capability.String("base-paths", "Comma-separated path prefixes for probing nonstandard LLM endpoint paths (e.g., /api,/proxy)").WithDefault(""),
	}
}

func (c *Capability) Match(_ capability.ExecutionContext, input capmodel.Port) error {
	if !strings.HasPrefix(input.Service, "http") {
		return fmt.Errorf("julius requires http/https service, got %q", input.Service)
	}
	return nil
}

func (c *Capability) Invoke(ctx capability.ExecutionContext, input capmodel.Port, output capability.Emitter) error {
	targetURL := fmt.Sprintf("%s://%s:%d", input.Service, input.Parent.DNS, input.Port)

	loadedProbes, err := probe.LoadProbesFromFS(probes.EmbeddedProbes, ".")
	if err != nil {
		return fmt.Errorf("loading probes: %w", err)
	}

	if basePaths, ok := ctx.Parameters.GetString("base-paths"); ok && basePaths != "" {
		paths := strings.Split(basePaths, ",")
		loadedProbes = probe.ExpandWithBasePaths(loadedProbes, paths)
	}

	sc := scanner.New(scanner.WithTimeout(30 * time.Second))
	results := sc.Scan(targetURL, loadedProbes, true)

	if len(results) == 0 {
		return nil
	}

	result := results[0]
	if result.Error != "" || result.Service == "" {
		return nil
	}

	parent := WebApplication{
		PrimaryURL: targetURL,
		Name:       result.Service,
		Status:     "A",
	}
	webpage := Webpage{
		URL:              targetURL,
		Type:             webpageType(result.Category),
		Service:          result.Service,
		GeneratorConfigs: toGeneratorConfigs(result.GeneratorConfigs),
		Specificity:      result.Specificity,
		Parent:           parent,
	}

	return output.Emit(parent, webpage)
}
