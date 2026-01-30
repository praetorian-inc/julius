//go:build augustus

package runner

var augustusFlag bool

func init() {
	probeCmd.Flags().BoolVar(&augustusFlag, "augustus", false, "Include Augustus generator configs in output")
}
