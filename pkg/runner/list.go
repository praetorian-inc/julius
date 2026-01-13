package runner

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/praetorian-inc/julius/pkg/probe"
	"github.com/praetorian-inc/julius/pkg/types"
	"github.com/praetorian-inc/julius/probes"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available probe definitions",
	Long: `List all probe definitions that are available for fingerprinting.
Shows the name, description, port hint, and number of probes for each definition.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	var loadedProbes []*types.ProbeDefinition
	var err error

	if probesDir != "" {
		loadedProbes, err = probe.LoadProbesFromDir(probesDir)
		if err != nil {
			return fmt.Errorf("loading probes from %s: %w", probesDir, err)
		}
	} else {
		loadedProbes, err = probe.LoadProbesFromFS(probes.EmbeddedProbes, ".")
		if err != nil {
			return fmt.Errorf("loading embedded probes: %w", err)
		}
	}

	if len(loadedProbes) == 0 {
		fmt.Println("No probe definitions found")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "DESCRIPTION", "PORT HINT", "PROBES", "CATEGORY"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, pd := range loadedProbes {
		portHint := fmt.Sprintf("%d", pd.PortHint)
		if pd.PortHint == 0 {
			portHint = "-"
		}
		probeCount := fmt.Sprintf("%d", len(pd.Probes))

		table.Append([]string{
			pd.Name,
			pd.Description,
			portHint,
			probeCount,
			pd.Category,
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d probe definitions\n", len(loadedProbes))
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
