package state

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the state list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflow states",
		Long: `List all workflow states in the Linear workspace.

Workflow states include: Todo, In Progress, Done, Canceled, etc.

Examples:
  linear state list
  linear state list --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 100, "Number of states to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	states, err := client.WorkflowStates(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list workflow states: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), states, true)
	case "table":
		if len(states.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No workflow states found")
			return nil
		}
		for _, state := range states.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", state.Name, state.Type)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
