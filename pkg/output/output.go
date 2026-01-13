package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/praetorian-inc/julius/pkg/types"
)

type TableWriter struct {
	writer io.Writer
}

func NewTableWriter(w io.Writer) types.OutputWriter {
	return &TableWriter{writer: w}
}

func (tw *TableWriter) Write(results []types.Result) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(tw.writer, "No matches found")
		return err
	}

	table := tablewriter.NewWriter(tw.writer)
	table.SetHeader([]string{"TARGET", "SERVICE", "CONFIDENCE", "MATCHED PROBE", "CATEGORY", "MODELS", "ERROR"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, result := range results {
		models := strings.Join(result.Models, ", ")

		table.Append([]string{
			result.Target,
			result.Service,
			result.Confidence,
			result.MatchedProbe,
			result.Category,
			models,
			result.Error,
		})
	}

	table.Render()
	return nil
}

type JSONWriter struct {
	writer io.Writer
}

func NewJSONWriter(w io.Writer) types.OutputWriter {
	return &JSONWriter{writer: w}
}

func (jw *JSONWriter) Write(results []types.Result) error {
	if len(results) == 0 {
		results = []types.Result{}
	}

	encoder := json.NewEncoder(jw.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func NewWriter(format string, w io.Writer) (types.OutputWriter, error) {
	switch format {
	case "table":
		return NewTableWriter(w), nil
	case "json":
		return NewJSONWriter(w), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}
