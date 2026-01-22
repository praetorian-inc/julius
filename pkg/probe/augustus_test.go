package probe

import (
	"testing"

	"github.com/praetorian-inc/augustus/pkg/generator"
	"github.com/praetorian-inc/julius/probes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllProbesValidAugustusConfig(t *testing.T) {
	loadedProbes, err := LoadProbesFromFS(probes.EmbeddedProbes, ".")
	require.NoError(t, err, "LoadProbesFromFS() should not error")

	for _, pd := range loadedProbes {
		t.Run(pd.Name, func(t *testing.T) {
			if pd.Augustus == nil {
				t.Skip("No Augustus config")
				return
			}

			configs := pd.BuildGeneratorConfigs("http://test.local:8080", []string{"test-model"})
			require.NotEmpty(t, configs, "Should produce at least one config")

			for i, cfg := range configs {
				_, err := generator.NewGenerator(cfg)
				assert.NoError(t, err, "Config %d should be valid", i)
			}
		})
	}
}

func TestBuildGeneratorConfigs(t *testing.T) {
	loadedProbes, err := LoadProbesFromFS(probes.EmbeddedProbes, ".")
	require.NoError(t, err)

	for _, pd := range loadedProbes {
		if pd.Augustus == nil {
			continue
		}

		t.Run(pd.Name, func(t *testing.T) {
			configs := pd.BuildGeneratorConfigs("http://test.local:8080", []string{"test-model"})
			require.Len(t, configs, 1)

			cfg := configs[0]
			assert.Equal(t, pd.Augustus.Generator, cfg.Type)
			assert.NotContains(t, cfg.Endpoint, "$TARGET", "Endpoint should have $TARGET resolved")
			assert.NotContains(t, cfg.Model, "$MODEL", "Model should have $MODEL resolved")
		})
	}
}
