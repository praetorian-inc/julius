package sdk

import (
	"fmt"
	"net"
	"strconv"
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
	scheme := "http"
	if strings.HasPrefix(input.Service, "https") {
		scheme = "https"
	}
	targetURL := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(input.Parent.DNS, strconv.Itoa(input.Port)))

	loadedProbes, err := probe.LoadProbesFromFS(probes.EmbeddedProbes, ".")
	if err != nil {
		return fmt.Errorf("loading probes: %w", err)
	}

	if basePaths, ok := ctx.Parameters.GetString("base-paths"); ok && basePaths != "" {
		raw := strings.Split(basePaths, ",")
		paths := make([]string, 0, len(raw))
		for _, p := range raw {
			if t := strings.TrimSpace(p); t != "" {
				paths = append(paths, t)
			}
		}
		if len(paths) > 0 {
			loadedProbes = probe.ExpandWithBasePaths(loadedProbes, paths)
		}
	}

	loadedProbes = probe.SortProbesByPortHint(loadedProbes, input.Port)

	sc := scanner.NewScanner(scanner.WithTimeout(30 * time.Second))
	results := sc.Scan(targetURL, loadedProbes, true)

	if len(results) == 0 {
		return nil
	}

	result := results[0]
	if result.Service == "" {
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
