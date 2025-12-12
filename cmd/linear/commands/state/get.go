package state

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the workflow state get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <state-id>",
		Short: "Get a single workflow state by ID",
		Long: `Get detailed information about a specific workflow state.

Retrieve full state details including name, type, position, color, and team information.
Workflow states represent stages in the issue lifecycle (e.g., "Todo", "In Progress", "Done").

Parameters:
  <state-id>: Workflow state UUID (required)

Output (--output=json):
  Returns JSON with: id, name, type, position, color, team

Examples:
  # Get state by UUID
  linear state get <state-uuid>

  # Get state with JSON output
  linear state get <state-uuid> --output=json

TIP: Use 'linear state list' to discover state IDs and names

Related Commands:
  - linear state list - List all workflow states
  - linear issue update --state=<name> - Set issue state`,
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

func runGet(cmd *cobra.Command, client *linear.Client, stateID string) error {
	ctx := context.Background()

	state, err := client.WorkflowState(ctx, stateID)
	if err != nil {
		return fmt.Errorf("failed to get workflow state: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), state, true, fieldSelector)
	case "table":
		// Simple table output for single state
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", state.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", state.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", state.Type)
		if state.Color != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Color:       %s\n", state.Color)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Position:    %.0f\n", state.Position)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
