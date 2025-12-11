package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single initiative by ID",
		Long: `Get detailed information about a specific strategic initiative.

Retrieve full initiative details including name, description, target dates, and progress.
Use this to track high-level strategic goals and their status.

Parameters:
  <id>: Initiative UUID (required)

Output (--output=json):
  Returns JSON with: id, name, description, targetDate, status

Examples:
  # Get initiative by UUID
  linear initiative get <initiative-uuid>

  # Get with JSON output
  linear initiative get <initiative-uuid> --output=json

TIP: Use 'linear initiative list' to discover initiative IDs

Related Commands:
  - linear initiative list - List all initiatives
  - linear project list - List projects within initiatives`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			initiative, err := client.Initiative(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get initiative: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				return formatter.FormatJSON(cmd.OutOrStdout(), initiative, true)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", initiative.Name)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
