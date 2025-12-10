package cycle

import (
	"context"
	"fmt"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

// NewGetCommand creates the cycle get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single cycle by ID",
		Long: `Get detailed information about a specific cycle.

Examples:
  linear cycle get <uuid>
  linear cycle get <uuid> --output=json`,
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

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, cycleID string) error {
	ctx := context.Background()

	cycle, err := client.Cycle(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), cycle, true)
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
