package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Validate probe definition files",
	Long: `Validate probe definition YAML files in a directory.
Checks each file for proper YAML syntax and required fields.

Example:
  julius validate ./probes`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	dir := args[0]

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", dir, err)
	}

	hasErrors := false
	validCount := 0
	errorCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
			continue
		}

		path := filepath.Join(dir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("ERROR: %s - failed to read: %v\n", filename, err)
			hasErrors = true
			errorCount++
			continue
		}

		p, err := probe.ParseProbe(data)
		if err != nil {
			fmt.Printf("ERROR: %s - invalid YAML: %v\n", filename, err)
			hasErrors = true
			errorCount++
			continue
		}

		errs := validateProbe(p)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Printf("ERROR: %s - %s\n", filename, e)
			}
			hasErrors = true
			errorCount++
			continue
		}

		fmt.Printf("OK: %s\n", filename)
		validCount++
	}

	fmt.Printf("\nValidation complete: %d valid, %d errors\n", validCount, errorCount)

	if hasErrors {
		return fmt.Errorf("validation failed with %d errors", errorCount)
	}

	return nil
}

func validateProbe(p *types.Probe) []string {
	var errors []string

	if p.Name == "" {
		errors = append(errors, "name is required")
	}

	if len(p.Requests) == 0 {
		errors = append(errors, "probe must have at least one request")
	}

	if p.Specificity < 0 || p.Specificity > 100 {
		errors = append(errors, fmt.Sprintf("specificity must be 0-100, got %d", p.Specificity))
	}

	for i, req := range p.Requests {
		if req.Path == "" {
			errors = append(errors, fmt.Sprintf("request %d: path is required", i))
		}
		if len(req.RawMatch) == 0 {
			errors = append(errors, fmt.Sprintf("request %d: at least one match rule is required", i))
		}
	}

	return errors
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
