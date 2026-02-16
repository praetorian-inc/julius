package runner

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/scanner"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/praetorian-inc/julius/probes"
	"github.com/spf13/cobra"
)

var (
	outputFormat       string
	probesDir          string
	timeout            int
	concurrency        int
	verbose            bool
	quiet              bool
	maxResponseSize    int64
	insecureSkipVerify bool
	caCertFile         string
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
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", scanner.DefaultConcurrency, "Maximum concurrent probe requests per target")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-match output")
	rootCmd.PersistentFlags().Int64Var(&maxResponseSize, "max-response-size", scanner.DefaultMaxResponseSize, "Maximum response body size in bytes (default 10MB)")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipVerify, "insecure", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringVar(&caCertFile, "ca-cert", "", "Path to custom CA certificate file")
}

func Run() error {
	return rootCmd.Execute()
}

// loadProbes loads probe definitions from the configured directory or embedded filesystem
func loadProbes() ([]*types.Probe, error) {
	if probesDir != "" {
		return probe.LoadProbesFromDir(probesDir)
	}
	return probe.LoadProbesFromFS(probes.EmbeddedProbes, ".")
}

// buildTLSConfig constructs a TLS configuration based on the configured flags.
// Returns nil when no TLS flags are set, preserving http.DefaultTransport behavior.
func buildTLSConfig() (*tls.Config, error) {
	if !insecureSkipVerify && caCertFile == "" {
		return nil, nil
	}

	if insecureSkipVerify && caCertFile != "" {
		fmt.Fprintln(os.Stderr, "Warning: --insecure overrides --ca-cert; custom CA certificate will be ignored")
	}

	tlsConfig := &tls.Config{}

	if insecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // User explicitly requested insecure mode for scanning targets with self-signed certs
	}

	if caCertFile != "" {
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
