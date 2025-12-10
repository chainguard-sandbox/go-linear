package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the project get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single project by ID",
		Long: `Get detailed information about a specific project.

Examples:
  linear project get <uuid>
  linear project get <uuid> --output=json`,
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

func runGet(cmd *cobra.Command, client *linear.Client, projectID string) error {
	ctx := context.Background()

	project, err := client.Project(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), project, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", project.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
