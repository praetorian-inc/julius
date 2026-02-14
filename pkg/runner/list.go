package runner

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available probe definitions",
	Long: `List all probe definitions that are available for fingerprinting.
Shows the name, description, port hint, and number of requests for each definition.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	loadedProbes, err := loadProbes()
	if err != nil {
		return fmt.Errorf("loading probes: %w", err)
	}

	if len(loadedProbes) == 0 {
		fmt.Println("No probe definitions found")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "DESCRIPTION", "PORT HINT", "REQUESTS", "SPECIFICITY", "CATEGORY"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, p := range loadedProbes {
		portHint := fmt.Sprintf("%d", p.PortHint)
		if p.PortHint == 0 {
			portHint = "-"
		}
		requestCount := fmt.Sprintf("%d", len(p.Requests))
		specificity := fmt.Sprintf("%d", p.GetSpecificity())

		table.Append([]string{
			p.Name,
			p.Description,
			portHint,
			requestCount,
			specificity,
			p.Category,
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d probe definitions\n", len(loadedProbes))
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
