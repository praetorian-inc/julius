package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/praetorian-inc/julius/pkg/probe"
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

		_, err = probe.ParseProbe(data)
		if err != nil {
			fmt.Printf("ERROR: %s - invalid YAML: %v\n", filename, err)
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

func init() {
	rootCmd.AddCommand(validateCmd)
}
