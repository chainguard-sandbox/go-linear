package cycle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the cycle get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single cycle by ID",
		Long: `Get detailed information about a specific development cycle (sprint).

Retrieve full cycle details including name, number, date range, and team information.

Parameters:
  <id>: Cycle UUID (required)

Output (--output=json):
  Returns JSON with: id, name, number, startsAt, endsAt, team

Examples:
  # Get cycle by UUID
  linear cycle get <uuid>

  # Get with JSON output
  linear cycle get <uuid> --output=json

TIP: Use 'linear cycle list' to discover cycle IDs

Related Commands:
  - linear cycle list - List all cycles
  - linear cycle update - Modify cycle details`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,name')")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, cycleID string) error {
	ctx := context.Background()

	cycle, err := client.Cycle(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), cycle, true, fieldSelector)
	case "table":
		if cycle.Name != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Name:   %s\n", *cycle.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Number: %.0f\n", cycle.Number)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
