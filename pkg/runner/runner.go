package runner

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	probesDir    string
	timeout      int
	verbose      bool
	quiet        bool
)

var rootCmd = &cobra.Command{
	Use:   "julius",
	Short: "Julius - LLM Service Fingerprinting Tool",
	Long: `Julius is a tool for fingerprinting LLM services by sending HTTP probes
and analyzing responses. It helps identify LLM platforms and available models.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, jsonl)")
	rootCmd.PersistentFlags().StringVarP(&probesDir, "probes-dir", "p", "", "Override probe definitions directory")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 5, "HTTP timeout in seconds")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-match output")
}

func Run() error {
	return rootCmd.Execute()
}
