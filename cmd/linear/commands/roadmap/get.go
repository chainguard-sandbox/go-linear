package roadmap

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the roadmap get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single roadmap by ID",
		Long: `Get detailed information about a specific roadmap.

Retrieve full roadmap details including name, description, and timeline configuration.

Parameters:
  <id>: Roadmap UUID (required)

Output (--output=json):
  Returns JSON with: id, name, description, createdAt

Examples:
  # Get roadmap by UUID
  linear roadmap get <uuid>

  # Get with JSON output
  linear roadmap get <uuid> --output=json

TIP: Use 'linear roadmap list' to discover roadmap IDs

Related Commands:
  - linear roadmap list - List all roadmaps
  - linear project list - List projects shown in roadmaps`,
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

func runGet(cmd *cobra.Command, client *linear.Client, roadmapID string) error {
	ctx := context.Background()

	roadmap, err := client.Roadmap(ctx, roadmapID)
	if err != nil {
		return fmt.Errorf("failed to get roadmap: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), roadmap, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", roadmap.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
