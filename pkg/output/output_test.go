package output

import (
	"bytes"
	"encoding/json"
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
			Target:       "https://api.openai.com",
			Service:      "OpenAI API",
			Confidence:   "high",
			MatchedProbe: "openai-completion",
			Category:     "llm",
		},
	}

	err := writer.Write(results)
	require.NoError(t, err, "Write should not fail")

	output := buf.String()
	assert.Contains(t, output, "TARGET")
	assert.Contains(t, output, "SERVICE")
	assert.Contains(t, output, "CONFIDENCE")

	assert.Contains(t, output, "https://api.openai.com")
	assert.Contains(t, output, "OpenAI API")
	assert.Contains(t, output, "high")
}

func TestTableWriter_WriteMultipleResults(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewTableWriter(buf)

	results := []types.Result{
		{
			Target:       "https://api.openai.com",
			Service:      "OpenAI API",
			Confidence:   "high",
			MatchedProbe: "openai-completion",
			Category:     "llm",
		},
		{
			Target:       "https://api.anthropic.com",
			Service:      "Anthropic API",
			Confidence:   "high",
			MatchedProbe: "anthropic-messages",
			Category:     "llm",
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
			Target:       "https://api.openai.com",
			Service:      "OpenAI API",
			Confidence:   "high",
			MatchedProbe: "openai-completion",
			Category:     "llm",
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
			Target:       "https://api.openai.com",
			Service:      "OpenAI API",
			Confidence:   "high",
			MatchedProbe: "openai-completion",
			Category:     "llm",
		},
		{
			Target:       "https://api.anthropic.com",
			Service:      "Anthropic API",
			Confidence:   "high",
			MatchedProbe: "anthropic-messages",
			Category:     "llm",
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
