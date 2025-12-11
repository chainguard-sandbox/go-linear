package cycle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the cycle list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all cycles",
		Long: `List all development cycles (sprints) in the Linear workspace.

Use this to browse current and past sprints, track cycle progress, or discover cycle IDs.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of cycles
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each cycle contains:
  - id: Cycle UUID
  - name: Cycle name (optional, may be null)
  - number: Cycle number
  - startsAt: Cycle start date
  - endsAt: Cycle end date
  - team: Associated team reference

Examples:
  # List all cycles
  linear cycle list

  # List with limit
  linear cycle list --limit=10

  # JSON output for parsing
  linear cycle list --output=json

Related Commands:
  - linear cycle get - Get single cycle details
  - linear cycle create - Create new cycle
  - linear issue list --cycle=<id> - Filter issues by cycle`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of cycles to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list cycles: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), cycles, true)
	case "table":
		if len(cycles.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No cycles found")
			return nil
		}
		for _, cycle := range cycles.Nodes {
			name := "Unnamed"
			if cycle.Name != nil {
				name = *cycle.Name
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%.0f - %.0f)\n", name, cycle.Number, cycle.Number)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
