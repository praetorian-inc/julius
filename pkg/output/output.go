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
	table.SetHeader([]string{"TARGET", "SERVICE", "SPECIFICITY", "CATEGORY", "MODELS", "ERROR"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, result := range results {
		models := strings.Join(result.Models, ", ")

		table.Append([]string{
			result.Target,
			result.Service,
			fmt.Sprintf("%d", result.Specificity),
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

type JSONLWriter struct {
	writer io.Writer
}

func NewJSONLWriter(w io.Writer) types.OutputWriter {
	return &JSONLWriter{writer: w}
}

func (jw *JSONLWriter) Write(results []types.Result) error {
	encoder := json.NewEncoder(jw.writer)
	for _, result := range results {
		if err := encoder.Encode(result); err != nil {
			return err
		}
	}
	return nil
}

func NewWriter(format string, w io.Writer) (types.OutputWriter, error) {
	switch format {
	case "table":
		return NewTableWriter(w), nil
	case "json":
		return NewJSONWriter(w), nil
	case "jsonl":
		return NewJSONLWriter(w), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}
