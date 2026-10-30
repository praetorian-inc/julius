package runner

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

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
	showBanner         bool
	noColor            bool
	maxResponseSize    int64
	insecureSkipVerify bool
	caCertFile         string
	proxyURL           string
	proxyAuth          string
)

var rootCmd = &cobra.Command{
	Use:   "julius",
	Short: "Julius - LLM Service Fingerprinting Tool",
	Long: `Julius is a tool for fingerprinting LLM services by sending HTTP probes
and analyzing responses. It helps identify LLM platforms and available models.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		useColor := isColorEnabled(noColor)
		if showBanner && !quiet && outputFormat == "table" {
			printBanner(useColor)
		}
	},
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
	rootCmd.PersistentFlags().StringVar(&proxyURL, "proxy", "", "Proxy URL (e.g. socks5://127.0.0.1:1080, http://127.0.0.1:8080)")
	rootCmd.PersistentFlags().StringVar(&proxyAuth, "proxy-auth", "", "Proxy authentication (username:password)")
	rootCmd.PersistentFlags().StringVar(&caCertFile, "ca-cert", "", "Path to custom CA certificate file")
	rootCmd.PersistentFlags().BoolVar(&showBanner, "banner", true, "Show ASCII banner")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
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

// buildProxyURL turns the --proxy/--proxy-auth flags into a URL for
// http.Transport. Returns nil when --proxy is unset, which leaves
// http.DefaultTransport's ProxyFromEnvironment behaviour intact so
// HTTP_PROXY/HTTPS_PROXY keep working as before.
//
// socks5h:// is accepted as an alias for socks5://. curl uses the h suffix to
// mean "resolve DNS at the proxy", which is what http.Transport does for
// socks5 anyway; Go does not recognise the scheme, so a socks5h URL copied
// from a working curl command would otherwise fail with an opaque
// "unsupported protocol scheme" at request time rather than at startup.
func buildProxyURL() (*url.URL, error) {
	if proxyURL == "" {
		if proxyAuth != "" {
			return nil, fmt.Errorf("--proxy-auth given without --proxy")
		}
		return nil, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parsing --proxy: %w", err)
	}
	if u.Scheme == "socks5h" {
		u.Scheme = "socks5"
	}
	switch u.Scheme {
	case "socks5", "http", "https":
	default:
		return nil, fmt.Errorf("unsupported --proxy scheme %q (want socks5, http or https)", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("--proxy %q has no host:port", proxyURL)
	}

	if proxyAuth != "" {
		user, pass, found := strings.Cut(proxyAuth, ":")
		if !found {
			return nil, fmt.Errorf("--proxy-auth must be username:password")
		}
		u.User = url.UserPassword(user, pass)
	}
	return u, nil
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
