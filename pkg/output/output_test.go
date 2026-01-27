package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTableWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewTableWriter(buf)

	require.NotNil(t, writer, "NewTableWriter should not return nil")
}

func TestTableWriter_WriteEmptyResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewTableWriter(buf)

	err := writer.Write([]types.Result{})
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Contains(t, output, "No matches found")
}

func TestTableWriter_WriteSingleResult(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewTableWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Contains(t, output, "TARGET")
	assert.Contains(t, output, "SERVICE")
	assert.Contains(t, output, "CONFIDENCE")
	assert.Contains(t, output, "SPECIFICITY")

	assert.Contains(t, output, "https://api.openai.com")
	assert.Contains(t, output, "OpenAI API")
	assert.Contains(t, output, "high")
	assert.Contains(t, output, "75")
}

func TestTableWriter_WriteMultipleResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewTableWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
		{
			Target:         "https://api.anthropic.com",
			Service:        "Anthropic API",
			Confidence:     "high",
			MatchedRequest: "anthropic-messages",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Contains(t, output, "OpenAI API")
	assert.Contains(t, output, "Anthropic API")
}

func TestNewJSONWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONWriter(buf)

	require.NotNil(t, writer, "NewJSONWriter should not return nil")
}

func TestJSONWriter_WriteEmptyResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONWriter(buf)

	err := writer.Write([]types.Result{})
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	var results []types.Result
	require.NoError(t, json.Unmarshal([]byte(output), &results), "Output should be valid JSON")

	assert.Empty(t, results, "Should return empty array")
	assert.Contains(t, output, "\n", "JSON should be indented")
}

func TestJSONWriter_WriteSingleResult(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	var parsed []types.Result
	require.NoError(t, json.Unmarshal([]byte(output), &parsed), "Output should be valid JSON")

	require.Len(t, parsed, 1, "Should return 1 result")
	assert.Equal(t, "https://api.openai.com", parsed[0].Target)
	assert.Equal(t, "OpenAI API", parsed[0].Service)
	assert.Equal(t, "high", parsed[0].Confidence)
}

func TestJSONWriter_WriteMultipleResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
		{
			Target:         "https://api.anthropic.com",
			Service:        "Anthropic API",
			Confidence:     "high",
			MatchedRequest: "anthropic-messages",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	var parsed []types.Result
	require.NoError(t, json.Unmarshal([]byte(output), &parsed), "Output should be valid JSON")

	require.Len(t, parsed, 2, "Should return 2 results")
	assert.Equal(t, "OpenAI API", parsed[0].Service)
	assert.Equal(t, "Anthropic API", parsed[1].Service)
}

func TestNewWriter_Table(t *testing.T) {
	buf := &bytes.Buffer{}
	writer, err := NewWriter("table", buf)

	require.NoError(t, err, "NewWriter should not fail")
	require.NotNil(t, writer, "NewWriter should not return nil")

	results := []types.Result{
		{
			Target:     "https://api.openai.com",
			Service:    "OpenAI API",
			Confidence: "high",
		},
	}

	err = writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Contains(t, output, "TARGET", "Output should be in table format")
}

func TestNewWriter_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	writer, err := NewWriter("json", buf)
	require.NoError(t, err, "NewWriter should not fail")

	require.NotNil(t, writer, "NewWriter should not return nil")

	results := []types.Result{
		{
			Target:     "https://test.com",
			Service:    "Test",
			Confidence: "high",
		},
	}

	err = writer.Write(results)
	require.NoError(t, err, "Write should not fail")
}

func TestNewWriter_UnknownFormat(t *testing.T) {
	_, err := NewWriter("xml", &bytes.Buffer{})
	assert.Error(t, err, "Should error for unknown format")
	assert.Contains(t, err.Error(), "unknown format")
}

func TestNewJSONLWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONLWriter(buf)

	require.NotNil(t, writer, "NewJSONLWriter should not return nil")
}

func TestJSONLWriter_WriteEmptyResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONLWriter(buf)

	err := writer.Write([]types.Result{})
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Empty(t, output, "Should return empty output for empty results")
}

func TestJSONLWriter_WriteSingleResult(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONLWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 1, "Should have 1 line")

	var parsed types.Result
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &parsed), "Line should be valid JSON")

	assert.Equal(t, "https://api.openai.com", parsed.Target)
	assert.Equal(t, "OpenAI API", parsed.Service)
	assert.Equal(t, "high", parsed.Confidence)
}

func TestJSONLWriter_WriteMultipleResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONLWriter(buf)

	results := []types.Result{
		{
			Target:         "https://api.openai.com",
			Service:        "OpenAI API",
			Confidence:     "high",
			MatchedRequest: "openai-completion",
			Category:       "llm",
			Specificity:    75,
		},
		{
			Target:         "https://api.anthropic.com",
			Service:        "Anthropic API",
			Confidence:     "high",
			MatchedRequest: "anthropic-messages",
			Category:       "llm",
			Specificity:    75,
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 2, "Should have 2 lines")

	var parsed1, parsed2 types.Result
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &parsed1), "Line 1 should be valid JSON")
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &parsed2), "Line 2 should be valid JSON")

	assert.Equal(t, "OpenAI API", parsed1.Service)
	assert.Equal(t, "Anthropic API", parsed2.Service)
}

func TestNewWriter_JSONL(t *testing.T) {
	buf := &bytes.Buffer{}
	writer, err := NewWriter("jsonl", buf)
	require.NoError(t, err, "NewWriter should not fail")

	require.NotNil(t, writer, "NewWriter should not return nil")

	results := []types.Result{
		{
			Target:     "https://test.com",
			Service:    "Test",
			Confidence: "high",
		},
	}

	err = writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 1, "Should have 1 line")
}

func TestTableWriterModelsAndError(t *testing.T) {
	tests := []struct {
		name           string
		result         types.Result
		expectInOutput []string
	}{
		{
			name: "with models",
			result: types.Result{
				Target:         "https://example.com",
				Service:        "ollama",
				Confidence:     "high",
				MatchedRequest: "/api/tags",
				Category:       "self-hosted",
				Specificity:    100,
				Models:         []string{"llama3.2:1b", "mistral:7b"},
			},
			expectInOutput: []string{"MODELS", "llama3.2:1b", "mistral:7b"},
		},
		{
			name: "with error",
			result: types.Result{
				Target:         "https://example.com",
				Service:        "openai",
				Confidence:     "high",
				MatchedRequest: "/v1/chat",
				Category:       "cloud-managed",
				Specificity:    75,
				Error:          "401 unauthorized",
			},
			expectInOutput: []string{"ERROR", "401 unauthorized"},
		},
		{
			name: "empty models and error",
			result: types.Result{
				Target:         "https://example.com",
				Service:        "test",
				Confidence:     "medium",
				MatchedRequest: "/health",
				Category:       "test",
				Specificity:    50,
			},
			expectInOutput: []string{"MODELS", "ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewTableWriter(&buf)

			err := writer.Write([]types.Result{tt.result})
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "expected %q in output", expected)
			}
		})
	}
}
