package runner

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/praetorian-inc/julius/pkg/output"
	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/scanner"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/spf13/cobra"
)

var (
	targetsFile  string
	augustusFlag bool
)

var probeCmd = &cobra.Command{
	Use:   "probe [targets...]",
	Short: "Probe targets to identify LLM services",
	Long: `Probe one or more targets to identify which LLM service they are using.

Targets can be specified in three ways:
  1. As command line arguments: julius probe https://api.example.com
  2. From a file: julius probe -f targets.txt
  3. From stdin: cat targets.txt | julius probe -

Examples:
  julius probe https://api.example.com
  julius probe -f targets.txt
  cat targets.txt | julius probe -
  julius probe https://api1.example.com https://api2.example.com`,
	RunE: runProbe,
}

func runProbe(cmd *cobra.Command, args []string) error {
	targets, err := loadTargets(args)
	if err != nil {
		return fmt.Errorf("loading targets: %w", err)
	}

	// Normalize all targets (trim whitespace, remove trailing slashes, add scheme if missing)
	targets = scanner.NormalizeTargets(targets)

	if len(targets) == 0 {
		return fmt.Errorf("no targets specified. Use --help for usage information")
	}

	loadedProbes, err := loadProbes()
	if err != nil {
		return fmt.Errorf("loading probes: %w", err)
	}

	if len(loadedProbes) == 0 {
		return fmt.Errorf("no probe definitions found")
	}

	tlsConfig, err := buildTLSConfig()
	if err != nil {
		return fmt.Errorf("building TLS config: %w", err)
	}

	timeoutDuration := time.Duration(timeout) * time.Second
	s := scanner.NewScanner(
		scanner.WithTimeout(timeoutDuration),
		scanner.WithConcurrency(concurrency),
		scanner.WithMaxResponseSize(maxResponseSize),
		scanner.WithTLSConfig(tlsConfig),
	)

	var allResults []types.Result

	for _, target := range targets {
		targetPort := scanner.ExtractPort(target)
		sortedProbes := probe.SortProbesByPortHint(loadedProbes, targetPort)

		results := s.Scan(target, sortedProbes, augustusFlag)
		if len(results) > 0 {
			allResults = append(allResults, results...)
		} else if !quiet {
			fmt.Fprintf(os.Stderr, "No match found for %s\n", target)
		}
	}

	writer, err := output.NewWriter(outputFormat, os.Stdout)
	if err != nil {
		return fmt.Errorf("creating output writer: %w", err)
	}

	if err := writer.Write(allResults); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func loadTargets(args []string) ([]string, error) {
	if len(args) == 0 {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, err
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return readTargetsFromReader(os.Stdin)
		}

		return nil, nil
	}

	if len(args) == 1 && args[0] == "-" {
		return readTargetsFromReader(os.Stdin)
	}

	if targetsFile != "" {
		f, err := os.Open(targetsFile)
		if err != nil {
			return nil, fmt.Errorf("opening targets file: %w", err)
		}
		defer f.Close()

		return readTargetsFromReader(f)
	}

	return args, nil
}

func readTargetsFromReader(r *os.File) ([]string, error) {
	var targets []string
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			targets = append(targets, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading targets: %w", err)
	}

	return targets, nil
}

func init() {
	rootCmd.AddCommand(probeCmd)
	probeCmd.Flags().StringVarP(&targetsFile, "file", "f", "", "Read targets from file")
	probeCmd.Flags().BoolVar(&augustusFlag, "augustus", false, "Include Augustus generator configs in output")
}
